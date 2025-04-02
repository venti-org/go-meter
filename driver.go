package gmeter

import (
	"sync"
)

type Driver struct {
	config    *DriverConfig
	stopped   bool
	requests  chan *Request
	meter     *Meter
	generator Generator[Request]
}

func NewDriver(config *DriverConfig) (*Driver, error) {
	generator, err := NewRequestGenerator(&config.RequestGeneratorConfig)
	if err != nil {
		return nil, err
	}
	return &Driver{
		config:    config,
		stopped:   false,
		requests:  make(chan *Request, 5),
		meter:     NewMeter(0),
		generator: generator,
	}, nil
}

func (driver *Driver) consume() error {
	defer close(driver.requests)
	allCount := driver.config.Concurrency * driver.config.ClientConfig.Count
	n := 0

	for {
		req, err := driver.generator.Generate()
		if err != nil {
			if driver.config.SkipError {
				continue
			}
			return err
		}
		if req == nil {
			break
		}
		if req.ID <= driver.config.Skip {
			continue
		}
		driver.requests <- req
		n += 1
		if n >= allCount {
			break
		}
	}
	return nil
}

func (driver *Driver) Run() error {
	var clients []*Client
	for i := range driver.config.Concurrency {
		if client, err := NewClient(i+1, &driver.config.ClientConfig); err != nil {
			return err
		} else {
			clients = append(clients, client)
		}
	}

	go func() {
		if err := driver.consume(); err != nil {
			ErrPrintln(err.Error())
		}
	}()

	wg := sync.WaitGroup{}
	for i := range clients {
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			client.Run(driver.requests)
		}(clients[i])
	}
	wg.Wait()
	for _, client := range clients {
		meter := client.GetMeter()
		driver.meter.Extend(meter)
		meter.Summary()
	}
	driver.meter.Summary()
	return nil
}

func (driver *Driver) Close() error {
	if driver.generator != nil {
		return driver.generator.Close()
	}
	return nil
}
