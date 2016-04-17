package ipam

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/asdine/storm"
	"github.com/mikioh/ipaddr"
	"github.com/tsuru/tsuru/log"
)

var store *storm.DB

type Range struct {
	ID string
}

type AvaliableIP struct {
	IP      string `storm:"id"`
	RangeId string
}

type AllocatedIP struct {
	IP      string `storm:"id"`
	RangeId string
}

func Init(s *storm.DB) error {
	store = s
	if err := s.Init(&Range{}); err != nil {
		log.Errorf("Range bucket creation failed: %v", err)
		return err
	}

	if err := s.Init(&AvaliableIP{}); err != nil {
		log.Errorf("AvaliableIP bucket creation failed: %v", err)
		return err
	}

	if err := s.Init(&AllocatedIP{}); err != nil {
		log.Errorf("AllocatedIP bucket creation failed: %v", err)
		return err
	}

	return nil
}

func InitRange(ipRange string) error {
	iprange := &Range{}
	err := store.One("ID", ipRange, iprange)

	if err != nil && err != storm.ErrNotFound {
		return err
	}

	if err == nil {
		logrus.Warnf("Range already initiated: %v", ipRange)
		return nil
	}

	iprange, err = newIpRange(ipRange)
	if err != nil {
		return err
	}

	err = store.Save(iprange)
	if err != nil {
		log.Errorf("InitRange failed: %v", err)
		return err
	}

	err = initAvaliableIPs(ipRange)
	if err != nil {
		return err
	}

	return nil
}

func Allocate() (string, error) {
	var ips []AvaliableIP

	if err := store.All(&ips); err != nil {
		return "", err
	}

	allocated := ips[0]

	if err := store.Remove(allocated); err != nil {
		return "", err
	}

	if err := store.Save(AllocatedIP{allocated.IP, allocated.RangeId}); err != nil {
		return "", err
	}

	return allocated.IP, nil
}

func Release(allocIp string) error {
	var ip AllocatedIP

	if err := store.One("IP", allocIp, &ip); err != nil {
		return err
	}

	if err := store.Remove(ip); err != nil {
		return err
	}

	if err := store.Save(AvaliableIP{ip.IP, ip.RangeId}); err != nil {
		return err
	}

	return nil
}

func initAvaliableIPs(ipRange string) error {
	c, err := ipaddr.Parse(ipRange)
	if err != nil {
		return err
	}

	for pos := c.Next(); pos != nil; pos = c.Next() {
		err := store.Save(AvaliableIP{
			IP:      pos.IP.String(),
			RangeId: ipRange,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func newIpRange(ipRange string) (*Range, error) {
	_, _, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, err
	}

	return &Range{
		ID: ipRange,
	}, nil
}
