package cluster

import "github.com/hashicorp/serf/serf"

type Cluster struct {
	Serf serf.Serf
}

func NewCluster() (*Cluster, error) {
	var c Cluster
	serf, err := setupSerf()
	if err != nil {
		return nil, err
	}

	c.Serf = *serf

	return &c, nil
}

func setupSerf() (*serf.Serf, error) {
	conf := serf.DefaultConfig()
	conf.Init()
	conf.Tags["role"] = "balancer"

	return serf.Create(conf)
}
