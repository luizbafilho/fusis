package ipvs

import "github.com/asdine/storm"

type IpvsStore struct {
	db *storm.DB
}

type IpvsKernel struct{}
