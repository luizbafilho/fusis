package health

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/luizbafilho/fusis/api/types"
)

//HealthCheck status
const (
	OK  = "ok"
	BAD = "bad"
)

type Check interface {
	Init(updateCh chan bool, dst types.Destination)
	Start()
	Stop()
	GetId() string
	GetStatus() string
}

type CheckTCP struct {
	Spec          types.CheckSpec
	DestinationID string
	Status        string

	tcp    string
	dialer *net.Dialer
	stopCh chan bool

	updateCh chan bool
}

func (c *CheckTCP) GetStatus() string {
	return c.Status
}

func (c *CheckTCP) Stop() {
	c.stopCh <- true
}

func (c *CheckTCP) GetId() string {
	return fmt.Sprintf("%s:%s", c.Spec.ServiceID, c.DestinationID)
}

// Start is used to start a TCP check.
// The check runs until stop is called
func (c *CheckTCP) Init(updateCh chan bool, dst types.Destination) {
	c.stopCh = make(chan bool)
	c.updateCh = updateCh

	if c.dialer == nil {
		// Create the socket dialer
		c.dialer = &net.Dialer{DualStack: true}

		// For long (>10s) interval checks the socket timeout is 10s, otherwise
		// the timeout is the interval. This means that a check *should* return
		// before the next check begins.
		if c.Spec.Timeout > 0 && c.Spec.Timeout < c.Spec.Interval {
			c.dialer.Timeout = c.Spec.Timeout
		} else if c.Spec.Interval < 10*time.Second {
			c.dialer.Timeout = c.Spec.Interval
		}
	}

	c.tcp = fmt.Sprintf("%s:%d", dst.Address, dst.Port)

	logrus.Debugf("[health-check] Init check on %s at %s. Interval=%s Timeout=%s", c.DestinationID, c.tcp, c.Spec.Interval, c.Spec.Timeout)
}

// check is invoked periodically to perform the TCP check
func (c *CheckTCP) check() {
	var currentStatus string
	previousStatus := c.Status

	conn, err := c.dialer.Dial(`tcp`, c.tcp)
	if err != nil {
		logrus.Warnf("[health-check] TCP Check is failing => %s", err)
		currentStatus = BAD
	} else {
		logrus.Debugf("[health-check] TCP check on '%v' at '%s' is passing", c.DestinationID, c.tcp)
		currentStatus = OK
		conn.Close()
	}

	if currentStatus != previousStatus {
		c.Status = currentStatus
		c.updateCh <- true
	}
}

func (c *CheckTCP) Start() {
	// Get the randomized initial pause time
	initialPauseTime := RandomStagger(c.Spec.Interval)
	next := time.After(initialPauseTime)
	for {
		select {
		case <-next:
			c.check()
			next = time.After(c.Spec.Interval)
		case <-c.stopCh:
			log.Printf("canceling check: %#v", c)
			return
		}
	}
}

// Returns a random stagger interval between 0 and the duration
func RandomStagger(intv time.Duration) time.Duration {
	if intv == 0 {
		return 0
	}
	return time.Duration(uint64(rand.Int63()) % uint64(intv))
}
