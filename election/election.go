package election

import (
	"context"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/luizbafilho/fusis/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Election struct {
	name      string
	election  *concurrency.Election
	electedCh chan bool
}

func New(config *config.BalancerConfig) (*Election, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   strings.Split(config.EtcdEndpoints, ","),
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, errors.Wrap(err, "[election] connection to etcd failed")
	}

	s, err := concurrency.NewSession(cli, concurrency.WithTTL(5))
	if err != nil {
		return nil, errors.Wrap(err, "[election] new session failed")
	}

	e := concurrency.NewElection(s, "/fusis-election")

	return &Election{
		name:      config.Name,
		election:  e,
		electedCh: make(chan bool),
	}, nil
}

func (e *Election) Run() chan bool {
	go e.observe()

	return e.electedCh
}

func (e *Election) IsLeader() bool {
	resp, err := e.election.Leader(context.TODO())
	if err != nil {
		return false
	}

	return string(resp.Kvs[0].Value) == e.name
}

func (e *Election) Resign() error {
	return e.election.Resign(context.TODO())
}

func (e *Election) observe() {
	if err := e.election.Campaign(context.Background(), e.name); err != nil {
		logrus.Error(errors.Wrap(err, "[election] campaign failed"))
	}

	// e.electedCh <- true

	for v := range e.election.Observe(context.TODO()) {
		if string(v.Kvs[0].Value) == e.name {
			e.electedCh <- true
		} else {
			e.electedCh <- false
			if err := e.election.Campaign(context.Background(), e.name); err != nil {
				logrus.Error(errors.Wrap(err, "[election] campaign failed"))
			}
		}
	}
}
