package state

import (
	"testing"

	"github.com/luizbafilho/fusis/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StateTestSuite struct {
	suite.Suite

	state State

	service     types.Service
	destination types.Destination
}

func TestStateTestSuite(t *testing.T) {
	suite.Run(t, new(StateTestSuite))
}

func (s *StateTestSuite) SetupTest() {
	var err error
	s.state, err = New()
	assert.Nil(s.T(), err)

	s.service = types.Service{
		Name:      "test",
		Address:   "10.0.1.1",
		Port:      80,
		Scheduler: "lc",
		Protocol:  "tcp",
	}

	s.destination = types.Destination{
		Name:      "test",
		Address:   "192.168.1.1",
		Port:      80,
		Mode:      "nat",
		Weight:    1,
		ServiceId: "test",
	}
}
