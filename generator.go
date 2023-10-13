package gmeter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Generator[T any] interface {
	Generate() (*T, error)
	io.Closer
}

type SimpleGenerator[T any] struct {
	callback func() (*T, error)
}

func NewSimpleGenerator[T any](callback func() (*T, error)) *SimpleGenerator[T] {
	return &SimpleGenerator[T]{
		callback: callback,
	}
}

func (generator *SimpleGenerator[T]) Generate() (*T, error) {
	return generator.callback()
}

func (generator *SimpleGenerator[T]) Close() error {
	return nil
}

type filterGenerator[T any] struct {
	filterFunc    func(*T) bool
	prevGenerator Generator[T]
}

func NewfilterGenerator[T any](prevGenerator Generator[T], filterFunc func(*T) bool) (*filterGenerator[T], error) {
	if prevGenerator == nil {
		return nil, fmt.Errorf("must set prev generator")
	}
	return &filterGenerator[T]{
		filterFunc:    filterFunc,
		prevGenerator: prevGenerator,
	}, nil
}

func (generator *filterGenerator[T]) Generate() (*T, error) {
	for {
		if value, err := generator.prevGenerator.Generate(); err != nil {
			return nil, err
		} else if value == nil {
			return nil, nil
		} else {
			if generator.filterFunc(value) {
				return value, nil
			}
		}
	}
}

func (generator *filterGenerator[T]) Close() error {
	return generator.prevGenerator.Close()
}

type MapGenerator[R any, T any] struct {
	mapFunc       func(*R) (*T, error)
	prevGenerator Generator[R]
}

func NewMapGenerator[R any, T any](prevGenerator Generator[R], mapFunc func(*R) (*T, error)) *MapGenerator[R, T] {
	return &MapGenerator[R, T]{
		mapFunc:       mapFunc,
		prevGenerator: prevGenerator,
	}
}

func (generator *MapGenerator[R, T]) Generate() (*T, error) {
	if value, err := generator.prevGenerator.Generate(); err != nil {
		return nil, err
	} else if value == nil {
		return nil, nil
	} else {
		return generator.mapFunc(value)
	}
}

func (generator *MapGenerator[R, T]) Close() error {
	return generator.prevGenerator.Close()
}

type FileGenerator struct {
	file   *os.File
	reader *bufio.Reader
}

func NewFileGenerator(path string) (*FileGenerator, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	return &FileGenerator{
		file:   file,
		reader: reader,
	}, nil
}

func (generator *FileGenerator) Generate() (*string, error) {
	reader := generator.reader
	if line, err := reader.ReadString('\n'); err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	} else {
		line = strings.TrimSuffix(line, "\n")
		return &line, nil
	}
}

func (generator *FileGenerator) Close() error {
	return generator.file.Close()
}

type UrlGenerator = Generator[string]
type BodyGenerator = Generator[io.Reader]

type RequestGenerator struct {
	Mode          int
	config        *RequestGeneratorConfig
	method        string
	urlGenerator  UrlGenerator
	bodyGenerator BodyGenerator
	index         int
	headers       []*Header
}

func NewJsonFileGenerator(path string) (Generator[map[string]interface{}], error) {
	if generator, err := NewFileGenerator(path); err != nil {
		return nil, err
	} else {
		return NewMapGenerator[string, map[string]interface{}](generator, func(b *string) (*map[string]interface{}, error) {
			body := make(map[string]interface{})
			if err := json.Unmarshal([]byte(*b), &body); err != nil {
				return nil, err
			}
			return &body, nil
		}), nil
	}
}

func readFileJson(path string) (map[string]interface{}, error) {
	if bytes, err := os.ReadFile(path); err != nil {
		return nil, err
	} else {
		var body map[string]interface{}
		if err := json.Unmarshal(bytes, &body); err != nil {
			return nil, err
		}
		return body, nil
	}
}

type Header struct {
	Key   string
	Value string
}

func parseHeaders(headers []string) []*Header {
	var heads []*Header
	for _, header := range headers {
		items := strings.SplitN(header, ":", 2)
		if len(items) != 2 {
			continue
		}
		head := &Header{
			Key:   strings.TrimSpace(items[0]),
			Value: strings.TrimSpace(items[1]),
		}
		if len(head.Key) != 0 {
			heads = append(heads, head)
		}
	}
	return heads
}

func NewRequestGenerator(config *RequestGeneratorConfig) (*RequestGenerator, error) {
	generator := &RequestGenerator{
		config:  config,
		method:  config.Method,
		headers: parseHeaders(config.Headers),
	}
	generator.bodyGenerator = NewSimpleGenerator(func() (*io.Reader, error) {
		var bodyReader io.Reader = strings.NewReader(config.Body)
		return &bodyReader, nil
	})
	if len(config.Url) != 0 {
		url := &config.Url
		generator.urlGenerator = NewSimpleGenerator(func() (*string, error) {
			return url, nil
		})
		if len(config.BodyPath) != 0 {
			if body, err := os.ReadFile(config.BodyPath); err != nil {
				return nil, err
			} else {
				generator.bodyGenerator = NewSimpleGenerator(func() (*io.Reader, error) {
					var reader io.Reader = bytes.NewReader(body)
					return &reader, nil
				})
			}
		} else if len(config.BodiesPath) != 0 {
			if len(config.ExtraJsonPath) != 0 {
				bodyGenerator, err := NewJsonFileGenerator(config.BodiesPath)
				if err != nil {
					return nil, err
				}
				if extraJson, err := readFileJson(config.ExtraJsonPath); err != nil {
					return nil, err
				} else {
					bodyGenerator = NewMapGenerator(bodyGenerator, func(r *map[string]interface{}) (*map[string]interface{}, error) {
						for key, value := range extraJson {
							(*r)[key] = value
						}
						return r, nil
					})
				}
				generator.bodyGenerator = NewMapGenerator(bodyGenerator, func(r *map[string]interface{}) (*io.Reader, error) {
					if b, err := json.Marshal(*r); err != nil {
						return nil, err
					} else {
						var reader io.Reader = bytes.NewReader(b)
						return &reader, nil
					}
				})
			} else {
				bodyGenerator, err := NewFileGenerator(config.BodiesPath)
				if err != nil {
					return nil, err
				}
				generator.bodyGenerator = NewMapGenerator((Generator[string])(bodyGenerator), func(s *string) (*io.Reader, error) {
					var reader io.Reader = strings.NewReader(*s)
					return &reader, nil
				})
			}
		}
	} else if len(config.UrlsPath) != 0 {
		if urlGenerator, err := NewFileGenerator(config.UrlsPath); err != nil {
			return nil, err
		} else {
			generator.urlGenerator = urlGenerator
		}
	} else {
		return nil, fmt.Errorf("must set url or urls-path")
	}
	return generator, nil
}

func (generator *RequestGenerator) nextUrl() (*string, error) {
	return generator.urlGenerator.Generate()
}

func (generator *RequestGenerator) nextBody() (*io.Reader, error) {
	return generator.bodyGenerator.Generate()
}

func (generator *RequestGenerator) Generate() (*Request, error) {
	var request *http.Request
	generator.index += 1
	var url *string
	var body *io.Reader
	var err error
	if url, err = generator.nextUrl(); err != nil {
	} else if body, err = generator.nextBody(); err != nil {
	} else {
		if url != nil && body != nil {
			request, err = http.NewRequest(generator.method, *url, *body)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("request (id %v) : %v", generator.index, err)
	}
	if request == nil {
		return nil, nil
	}
	for _, header := range generator.headers {
		request.Header.Add(header.Key, header.Value)
	}
	return &Request{
		ID:  generator.index,
		Req: request,
	}, nil
}

func (generator *RequestGenerator) Close() error {
	var errs []error
	if generator.urlGenerator != nil {
		if err := generator.urlGenerator.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if generator.bodyGenerator != nil {
		if err := generator.bodyGenerator.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return GainError(errs)
}
