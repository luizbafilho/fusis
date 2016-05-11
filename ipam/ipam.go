package ipam

import (
	"net"
	"sync"
	"time"

	"github.com/luizbafilho/fusis/ipvs"
	"github.com/mikioh/ipaddr"
)

var rangeCursor *ipaddr.Cursor
var allocatedIps []net.IP
var mutex sync.Mutex

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
		containsAllocated, err := allocatedContains(pos.IP.String())
		if err != nil {
			return "", err
		}

		containsAssigned, err := assignedContains(pos.IP.String())
		if err != nil {
			return "", err
		}

		if !containsAllocated && !containsAssigned {
			allocateIP(pos.IP)

			rangeCursor.Set(rangeCursor.First())
			return pos.IP.String(), nil
		}
	}

	return "", nil
}

//Release releases a allocated IP
func Release(allocIP string) {
	removeAllocatedIP(allocIP)
}

func allocateIP(ip net.IP) {
	mutex.Lock()
	allocatedIps = append(allocatedIps, ip)
	mutex.Unlock()

	go cleanAllocation(ip.String())
}

func cleanAllocation(ip string) {
	time.Sleep(time.Second * 30)
	removeAllocatedIP(ip)
}

func removeAllocatedIP(ip string) {
	mutex.Lock()
	var index int

	for i, a := range allocatedIps {
		if a.String() == ip {
			index = i
		}
	}

	allocatedIps = append(allocatedIps[:index], allocatedIps[index+1:]...)

	mutex.Unlock()
}

func assignedContains(e string) (bool, error) {
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

func allocatedContains(e string) (bool, error) {
	mutex.Lock()
	for _, a := range allocatedIps {
		if a.String() == e {
			return true, nil
		}
	}
	mutex.Unlock()

	return false, nil
}
