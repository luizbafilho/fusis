package health

import (
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
)

//HealthCheck status
const (
	OK  = "ok"
	BAD = "bad"
)

type CheckNotifier interface {
	NotifyCheckUpdate() error
}

// Check is used to periodically make an TCP/UDP connection to
// determine the health of a given check.
// The check is OK if the connection succeeds
// The check is BAD if the connection returns an error
type Check struct {
	UpdatesCh     chan Check `json:"-"`
	Status        string
	DestinationID string
	TCP           string

	Interval time.Duration
	Timeout  time.Duration

	dialer   *net.Dialer
	stop     bool
	stopCh   chan struct{}
	stopLock sync.Mutex
}

// Start is used to start a TCP check.
// The check runs until stop is called
func (c *Check) Start() {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()

	if c.dialer == nil {
		// Create the socket dialer
		c.dialer = &net.Dialer{DualStack: true}

		// For long (>10s) interval checks the socket timeout is 10s, otherwise
		// the timeout is the interval. This means that a check *should* return
		// before the next check begins.
		if c.Timeout > 0 && c.Timeout < c.Interval {
			c.dialer.Timeout = c.Timeout
		} else if c.Interval < 10*time.Second {
			c.dialer.Timeout = c.Interval
		}
	}

	c.stop = false
	c.stopCh = make(chan struct{})
	go c.run()
}

// Stop is used to stop a TCP check.
func (c *Check) Stop() {
	c.stopLock.Lock()
	defer c.stopLock.Unlock()
	if !c.stop {
		c.stop = true
		close(c.stopCh)
	}
}

// run is invoked by a goroutine to run until Stop() is called
func (c *Check) run() {
	// Get the randomized initial pause time
	initialPauseTime := RandomStagger(c.Interval)
	logrus.Printf("[DEBUG] agent: pausing %v before first socket connection of %s", initialPauseTime, c.TCP)
	next := time.After(initialPauseTime)
	for {
		select {
		case <-next:
			c.check()
			next = time.After(c.Interval)
		case <-c.stopCh:
			return
		}
	}
}

// check is invoked periodically to perform the TCP check
func (c *Check) check() {
	var currentStatus string
	previousStatus := c.Status

	conn, err := c.dialer.Dial(`tcp`, c.TCP)
	if err != nil {
		logrus.Printf("[WARN] agent: socket connection failed '%s': %s", c.TCP, err)
		currentStatus = BAD
	} else {
		logrus.Printf("[DEBUG] agent: check '%v' is passing", c.DestinationID)
		currentStatus = OK
		conn.Close()
	}

	if currentStatus != previousStatus {
		c.Status = currentStatus
		c.UpdatesCh <- *c
	}
}

// Returns a random stagger interval between 0 and the duration
func RandomStagger(intv time.Duration) time.Duration {
	if intv == 0 {
		return 0
	}
	return time.Duration(uint64(rand.Int63()) % uint64(intv))
}
