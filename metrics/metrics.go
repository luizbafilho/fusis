package metrics

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/state"
	"github.com/pkg/errors"
)

type Collector interface {
	Monitor()
}

type Publisher interface {
	PublishServiceStats(stats *ipvs.Service) error
	Close() error
}

type Metrics struct {
	config    *config.BalancerConfig
	publisher Publisher
}

func NewMetrics(state state.State, config *config.BalancerConfig) Collector {
	return &Metrics{
		config: config,
	}
}

func (m *Metrics) Monitor() {
	if m.config.Metrics.Publisher == "" {
		return
	}

	err := m.InitPublisher()
	if err != nil {
		logrus.Warnf("Monitoring failed. %v", err)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for _ = range ticker.C {
		err := m.collect()
		if err != nil {
			ticker.Stop()
			logrus.Warnf("Collecting metrics failed. Stoping monitoring statistics. %v", err)
		}
	}

	//Closing publish connection
	m.publisher.Close()
}

func (m *Metrics) InitPublisher() error {
	var err error
	switch m.config.Metrics.Publisher {
	case "logstash":
		m.publisher, err = NewLogstashPublisher(m.config)
	}

	if err != nil {
		return err
	}

	return nil
}

func (m Metrics) collect() error {
	services, err := ipvs.GetServices()
	if err != nil {
		return errors.Wrap(err, "ipvs.GetServices() failed when collecting metrics")
	}

	for _, s := range services {
		if err := m.publisher.PublishServiceStats(s); err != nil {
			return err
		}
	}

	return nil
}
