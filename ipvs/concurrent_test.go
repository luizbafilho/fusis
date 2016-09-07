package ipvs

import (
	"log"
	"sync"
	"syscall"

	gipvs "github.com/google/seesaw/ipvs"
	"github.com/k0kubun/pp"
	"github.com/mikioh/ipaddr"
	. "gopkg.in/check.v1"
)

func (s *IpvsSuite) TestConcurrent(c *C) {
	gipvs.Init()

	cursor := getRange()

	count := 0
	var mutex = &sync.Mutex{}

	for pos := cursor.Next(); pos != nil; pos = cursor.Next() {
		mutex.Lock()
		if count >= 10 {
			break
		}
		mutex.Unlock()

		addr := pos.IP
		go func() {
			service := gipvs.Service{
				Address:   addr,
				Protocol:  syscall.IPPROTO_TCP,
				Port:      54321,
				Scheduler: "wlc",
				Flags:     0,
				Timeout:   100000,
			}

			mutex.Lock()
			err := gipvs.AddService(service)
			mutex.Unlock()
			if err != nil {
				pp.Println(err)
			}

			mutex.Lock()
			count = count + 1
			mutex.Unlock()
		}()
	}
}

func getRange() *ipaddr.Cursor {
	var rangeCursor *ipaddr.Cursor
	var err error

	rangeCursor, err = ipaddr.Parse("100.0.0.0/16")
	if err != nil {
		log.Fatal(err)
	}

	return rangeCursor

}
