package gmeter

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	id     int
	client *http.Client
	config *ClientConfig
	meter  *Meter
}

func NewClient(id int, config *ClientConfig) (*Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if len(config.Proxy) != 0 {
		if proxyUrl, err := url.Parse(config.Proxy); err != nil {
			return nil, err
		} else {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	return &Client{
		id: id,
		client: &http.Client{
			Transport: transport,
		},
		config: config,
		meter:  NewMeter(id),
	}, nil
}

func (Client *Client) GetMeter() *Meter {
	return Client.meter
}

func (client *Client) Run(requests chan *Request) {
	for i := 0; i < client.config.Count; i++ {
		request := <-requests
		if request == nil {
			break
		}
		client.meter.Start()
		start := time.Now()
		response, err := client.client.Do(request.Req)
		res := NewResponse(request, response, err)
		res.Cost = time.Since(start).Milliseconds()
		client.meter.Finish(res)
	}
}
