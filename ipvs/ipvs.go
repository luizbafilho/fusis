package ipvs

import "github.com/luizbafilho/fusis/db"

var Store *IpvsStore
var Kernel *IpvsKernel

func Init(s *db.DB) error {
	if err := initStore(s); err != nil {
		return err
	}

	if err := initKernel(); err != nil {
		return err
	}

	return nil
}
