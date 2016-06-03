package none

import (
	"os"
	"testing"

	"github.com/asdine/storm"
	"github.com/luizbafilho/fusis/ipvs"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type IpamSuite struct {
	db *storm.DB
}

var _ = Suite(&IpamSuite{})

func (s *IpamSuite) SetUpSuite(c *C) {
	s.db, _ = db.New("fusis_test.db")

	ipvs.InitStore(s.db)

	if err := Init("192.168.0.0/28"); err != nil {
		panic(err)
	}
}

func (s *IpamSuite) TearDownSuite(c *C) {
	os.Remove("fusis_test.db")
}

func (s *IpamSuite) TestIpAllocation(c *C) {
	service := &ipvs.Service{
		Id:   "test",
		Host: "192.168.0.1",
	}
	s.db.Save(service)

	ip, err := Allocate()
	c.Assert(err, IsNil)
	c.Assert(ip, DeepEquals, "192.168.0.2")

	service = &ipvs.Service{
		Id:   "test2",
		Host: "192.168.0.2",
	}
	s.db.Save(service)

	ip, err = Allocate()
	c.Assert(err, IsNil)
	c.Assert(ip, DeepEquals, "192.168.0.3")

	s.db.Remove(service)

	ip, err = Allocate()
	c.Assert(err, IsNil)
	c.Assert(ip, DeepEquals, "192.168.0.2")
}
