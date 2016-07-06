package ipvs

import (
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	logrus_logstash "github.com/bshuster-repo/logrus-logstash-hook"

	"github.com/luizbafilho/fusis/api/types"
)

func RouterLog(statsLog *logrus.Logger, tick time.Time, s types.Service) {

	hosts := []string{}
	for _, dst := range s.Destinations {
		hosts = append(hosts, dst.Host)
	}

	statsLog.WithFields(logrus.Fields{
		"time":     tick,
		"service":  s.Name,
		"Protocol": s.Protocol,
		"Port":     s.Port,
		"hosts":    strings.Join(hosts, ","),
		"client":   "fusis",
	}).Info("Fusis router stats")
}

func LogStash() {

	// TODO: Generalise this options to come from fusis.json.
	// At engine.go verify the log type and set the logger

	logger := logrus.New()
	PROTOCOL := "udp"
	HOST := "logstash.video.dev.globoi.com"
	PORT := "8515"
	url := []string{HOST, ":", PORT}
	hook, err := logrus_logstash.NewHook(PROTOCOL, strings.Join(url, ""), "Fusis")
	if err != nil {
		logger.Fatal(err)
	}

	logger.Hooks.Add(hook)
}
