package ipvs

var Store *IpvsStore
var Kernel *IpvsKernel

func Init() error {
	// if err := InitStore(); err != nil {
	// 	return err
	// }

	if err := initKernel(); err != nil {
		return err
	}

	return nil
}
