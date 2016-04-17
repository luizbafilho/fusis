package ipvs

import "github.com/asdine/storm"

var Store *IpvsStore
var Kernel *IpvsKernel

func Init(s *storm.DB) error {
	if err := initStore(s); err != nil {
		return err
	}

	if err := initKernel(); err != nil {
		return err
	}

	return nil
}
