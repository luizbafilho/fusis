package ipam

import (
	"os"
	"testing"

	"github.com/asdine/storm"
	"github.com/luizbafilho/fusis/db"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpamSuite struct {
	db      *storm.DB
	ipRange string
}

var _ = Suite(&IpamSuite{})

func (s *IpamSuite) SetUpSuite(c *C) {
	var stormDB *storm.DB
	stormDB, err := db.New("fusis_test.db")
	if err != nil {
		panic(err)
	}

	s.db = stormDB
	s.ipRange = "192.168.1.0/28"

	Init(s.db)
}

func (s *IpamSuite) TearDownSuite(c *C) {
	os.Remove("fusis_test.db")
}

func (s *IpamSuite) SetUpTest(c *C) {
	InitRange(s.ipRange)
}

func (s *IpamSuite) TearDownTest(c *C) {
	var rs []Range
	store.All(&rs)
	for _, r := range rs {
		store.Remove(r)
	}

	var aips []AvaliableIP
	store.All(&aips)

	for _, a := range aips {
		store.Remove(a)
	}

	var alips []AllocatedIP
	store.All(&alips)

	for _, a := range alips {
		store.Remove(a)
	}
}

func (s *IpamSuite) TestRangeInitialization(c *C) {
	savedRange := Range{}
	s.db.One("ID", s.ipRange, &savedRange)

	c.Assert(savedRange, DeepEquals, Range{s.ipRange})

	var avaliableIps []AvaliableIP
	s.db.All(&avaliableIps)

	c.Assert(avaliableIps[0], DeepEquals, AvaliableIP{"192.168.1.1", "192.168.1.0/28"})
}

func (s *IpamSuite) TestRangeInitializationValidation(c *C) {
	ip, _ := Allocate()
	InitRange(s.ipRange)

	var aip AvaliableIP
	err := s.db.One("IP", ip, &aip)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), DeepEquals, "not found")
}

func (s *IpamSuite) TestIpAllocation(c *C) {
	ip, err := Allocate()
	c.Assert(err, IsNil)
	c.Assert(ip, DeepEquals, "192.168.1.1")

	var allocatedIps []AllocatedIP
	s.db.All(&allocatedIps)
	c.Assert(allocatedIps[0], DeepEquals, AllocatedIP{"192.168.1.1", "192.168.1.0/28"})

	var avaliableIp AvaliableIP
	err = s.db.One("IP", ip, &avaliableIp)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), DeepEquals, "not found")
}

func (s *IpamSuite) TestIpRelease(c *C) {
	ip := "192.168.1.1"
	Release(ip)

	var allocatedIps []AllocatedIP
	s.db.All(&allocatedIps)

	c.Assert(len(allocatedIps), DeepEquals, 0)

	var avaliableIp AvaliableIP
	s.db.One("IP", ip, &avaliableIp)

	c.Assert(avaliableIp, DeepEquals, AvaliableIP{ip, s.ipRange})
}
