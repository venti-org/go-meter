package gmeter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type Driver struct {
	config  *DriverConfig
	stopped bool
	bodys   chan map[string]interface{}
	meter   *Meter
}

func NewDriver(config *DriverConfig) *Driver {
	return &Driver{
		config:  config,
		stopped: false,
		bodys:   make(chan map[string]interface{}, 5),
		meter:   NewMeter(0),
	}
}

func (driver *Driver) consume() error {
	defer close(driver.bodys)
	extractJson := make(map[string]interface{})
	if len(driver.config.ExtraJsonPath) != 0 {
		if data, err := os.ReadFile(driver.config.ExtraJsonPath); err != nil {
			return err
		} else if err = json.Unmarshal(data, &extractJson); err != nil {
			return err
		}
	}
	if len(driver.config.Path) == 0 {
		return fmt.Errorf("must set path")
	}
	jsonFile, err := os.Open(driver.config.Path)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(jsonFile)
	allCount := driver.config.Concurrency * driver.config.Count
	n := 0
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		body := make(map[string]interface{})
		if err := json.Unmarshal(line, &body); err != nil {
            if driver.config.SkipError {
                continue
            }
			return err
		}
		for k, v := range extractJson {
			body[k] = v
		}
		driver.bodys <- body
		n += 1
		if n >= allCount {
			break
		}
	}

	return jsonFile.Close()
}

func (driver *Driver) Run() {
	go func() {
		if err := driver.consume(); err != nil {
			fmt.Println(err)
		}
	}()

	var clients []*Client
	clientConfig := &ClientConfig{
		Count: driver.config.Count,
		Api:   driver.config.Api,
	}
	for i := 0; i < driver.config.Concurrency; i++ {
		clients = append(clients, NewClient(i+1, clientConfig))
	}
	wg := sync.WaitGroup{}
	for i := 0; i < len(clients); i++ {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			client.Run(driver.bodys)
		}(clients[i])
	}
	wg.Wait()
	for i := 0; i < len(clients); i++ {
		meter := clients[i].GetMeter()
		driver.meter.Extend(meter)
		meter.Summary()
		fmt.Println()
	}
	driver.meter.Summary()
}
