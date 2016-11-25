package health

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/luizbafilho/fusis/api/types"
)

//HealthCheck status
const (
	OK  = "ok"
	BAD = "bad"
)

type Check interface {
	Init(updateCh chan bool)
	Start()
	Stop()
	GetId() string
}

type CheckTCP struct {
	Spec types.CheckSpec

	DestinationID string

	dialer *net.Dialer
	stopCh chan bool

	updateCh chan bool
}

func (c *CheckTCP) Stop() {
	c.stopCh <- true
}

func (c *CheckTCP) GetId() string {
	return fmt.Sprintf("%s:%s", c.Spec.ServiceID, c.DestinationID)
}

// Start is used to start a TCP check.
// The check runs until stop is called
func (c *CheckTCP) Init(updateCh chan bool) {
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
}

// check is invoked periodically to perform the TCP check
func (c *CheckTCP) check() {
	log.Printf("Check Running: %#v", c)
	// var currentStatus string
	// previousStatus := c.Status
	//
	// conn, err := c.dialer.Dial(`tcp`, c.TCP)
	// if err != nil {
	// 	logrus.Printf("[WARN] agent: socket connection failed '%s': %s", c.TCP, err)
	// 	currentStatus = BAD
	// } else {
	// 	logrus.Printf("[DEBUG] agent: check '%v' is passing", c.DestinationID)
	// 	currentStatus = OK
	// 	conn.Close()
	// }
	//
	// if currentStatus != previousStatus {
	// 	c.Status = currentStatus
	// 	c.UpdatesCh <- *c
	// }
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
