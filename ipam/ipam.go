package ipam

import (
	"net"

	"github.com/mikioh/ipaddr"
)

var store *store.StoreBolt

type IpRange struct {
	ID    string
	Range *ipaddr.Prefix
}

func Init(store *store.StoreBolt) {
	store = store
}

func InitRange(ipRange string) error {
	_, err := newIpRange(ipRange)
	if err != nil {
		return err
	}
	return nil
}

func newIpRange(ipRange string) (*IpRange, error) {
	_, net, err := net.ParseCIDR(ipRange)
	if err != nil {
		return nil, err
	}

	return &IpRange{
		ID:    ipRange,
		Range: ipaddr.NewPrefix(net),
	}, nil
}
