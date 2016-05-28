package ipam

import (
	"github.com/luizbafilho/fusis/ipvs"
	"github.com/mikioh/ipaddr"
)

var rangeCursor *ipaddr.Cursor

//Init initilizes ipam module
func Init(iprange string) error {
	var err error
	rangeCursor, err = ipaddr.Parse(iprange)
	if err != nil {
		return err
	}

	return nil
}

//Allocate allocates a new avaliable ip
func Allocate() (string, error) {
	for pos := rangeCursor.Next(); pos != nil; pos = rangeCursor.Next() {
		assigned, err := ipIsAssigned(pos.IP.String())
		if err != nil {
			return "", err
		}

		if !assigned {
			rangeCursor.Set(rangeCursor.First())
			return pos.IP.String(), nil
		}
	}

	return "", nil
}

//Release releases a allocated IP
func Release(allocIP string) {}

func ipIsAssigned(e string) (bool, error) {
	services, err := ipvs.Store.GetServices()
	if err != nil {
		return false, err
	}

	for _, a := range *services {
		if a.Host == e {
			return true, nil
		}

	}
	return false, nil
}
