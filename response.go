package gmeter

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"unicode/utf8"

	"golang.org/x/net/html/charset"
)

type Response struct {
	RequestUrl       string
	ResponseUrl      string
	ResponseMimeType string
	Error            error
	StatusCode       int
	Body             any
	BodyError        error
	Cost             int64
	ID               int
}

func NewResponse(request *Request, response *http.Response, err error) *Response {
	res := &Response{
		Error:      err,
		RequestUrl: request.Req.URL.String(),
		ID:         request.ID,
	}
	if response != nil {
		defer response.Body.Close()
		res.StatusCode = response.StatusCode
		res.ResponseUrl = response.Request.URL.String()
		contentType := response.Header.Get("Content-Type")
		mediatype, _, _ := mime.ParseMediaType(contentType)
		if strings.HasPrefix(mediatype, "text/") {
			if body, err := charset.NewReader(response.Body, contentType); err != nil {
				res.BodyError = err
			} else {
				content, err := io.ReadAll(body)
				if err != nil {
					res.BodyError = err
				} else {
					res.Body = string(content)
				}
			}
		} else if mediatype == "application/json" {
			result := make(map[string]any)
			if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
				res.BodyError = err
			} else {
				res.Body = result
			}
		} else {
			if body, err := io.ReadAll(response.Body); err != nil {
				res.BodyError = err
			} else {
				if utf8.Valid(body) {
					res.Body = string(body)
				} else {
					res.Body = body
				}
			}
		}
	}
	return res
}

func (res *Response) DefaultJson() map[string]any {
	result := make(map[string]any)
	result["url"] = res.RequestUrl
	result["cost"] = res.Cost
	result["id"] = res.ID
	return result
}

func (res *Response) ErrorJson() (map[string]any, error) {
	if res.Error == nil {
		return nil, fmt.Errorf("Response error is nil")
	} else {
		result := res.DefaultJson()
		result["code"] = 1
		result["error"] = res.Error.Error()
		return result, nil
	}
}

func (res *Response) SuccessJson() (map[string]any, error) {
	if res.Error != nil {
		return nil, fmt.Errorf("Response error is not nil")
	} else {
		result := res.DefaultJson()
		result["code"] = 0
		result["response_url"] = res.ResponseUrl
		result["status_code"] = res.StatusCode
		result["body"] = res.Body
		if res.BodyError != nil {
			result["body_error"] = res.BodyError.Error()
		}
		return result, nil
	}
}

func (res *Response) String() string {
	var result map[string]any
	var err error
	if res.Error != nil {
		result, err = res.ErrorJson()
	} else {
		result, err = res.SuccessJson()
	}
	if err == nil {
		var s []byte
		if s, err = json.Marshal(result); err == nil {
			return string(s)
		}
	}
	result = res.DefaultJson()
	result["code"] = 1
	result["error"] = err.Error()
	if s, err := json.Marshal(result); err == nil {
		return string(s)
	} else {
		return err.Error()
	}
}
