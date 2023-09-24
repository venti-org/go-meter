package gmeter

import (
	"bytes"
	"encoding/json"
	"net/http"
)

type Client struct {
	id     int
	client *http.Client
	config *ClientConfig
	meter  *Meter
}

func NewClient(id int, config *ClientConfig) *Client {
	return &Client{
		id:     id,
		client: &http.Client{},
		config: config,
		meter:  NewMeter(id),
	}
}

func (Client *Client) GetMeter() *Meter {
	return Client.meter
}

func (client *Client) Run(bodys chan map[string]interface{}) {
	for i := 0; i < client.config.Count; i++ {
		body := <-bodys
		if body == nil {
			break
		}
		client.meter.Start()
		result, err := client.post(client.config.Api, body)
		client.meter.Finish(result, err)
	}
}

func (client *Client) post(api string, body map[string]interface{}) (map[string]interface{}, error) {
	if data, err := json.Marshal(body); err != nil {
		return nil, err
	} else if resp, err := client.client.Post(api, "application/json", bytes.NewReader(data)); err != nil {
		return nil, err
	} else {
		defer resp.Body.Close()
		result := make(map[string]interface{})
		if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return result, nil
	}
}
