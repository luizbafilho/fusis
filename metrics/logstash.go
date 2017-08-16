package metrics

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/luizbafilho/fusis/config"
	"github.com/mqliang/libipvs"
	"github.com/pkg/errors"
)

type Logstash struct {
	conn   net.Conn
	extras map[string]string
}

func NewLogstashPublisher(config *config.BalancerConfig) (Publisher, error) {
	address := fmt.Sprintf("%s:%d", config.Metrics.Params["host"], config.Metrics.Params["port"])
	conn, err := net.Dial("udp", address)
	if err != nil {
		return nil, errors.Wrap(err, "logstash connection failed")
	}

	return &Logstash{
		conn:   conn,
		extras: config.Metrics.Extras,
	}, nil
}

//Closes logstash connection
func (l Logstash) Close() error {
	return l.conn.Close()
}

func (l Logstash) PublishServiceStats(svc *libipvs.Service) error {
	data := map[string]interface{}{}
	for k, v := range l.extras {
		data[k] = v
	}

	data["service_address"] = svc.Address.String()
	data["service_protocol"] = svc.Protocol.String()
	data["service_port"] = svc.Port

	// for _, d := range libipvs.ListDestinations(svc) {
	// 	data["destination_address"] = d.Address.String()
	// 	data["destination_port"] = d.Port
	// 	data["@timestamp"] = time.Now().Unix()
	//
	// 	stats := structs.New(d.Statistics)
	// 	for _, f := range stats.Fields() {
	// 		if f.IsEmbedded() {
	// 			for _, s := range f.Fields() {
	// 				data["name"] = s.Name()
	// 				data["value"] = s.Value()
	//
	// 				if err := l.publish(data); err != nil {
	// 					return err
	// 				}
	// 			}
	// 			continue
	// 		}
	//
	// 		data["name"] = f.Name()
	// 		data["value"] = f.Value()
	//
	// 		if err := l.publish(data); err != nil {
	// 			return err
	// 		}
	// 	}
	// }

	return nil
}

func (l Logstash) publish(data map[string]interface{}) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Metric data marshaling failed")
	}

	_, err = l.conn.Write(bytes)
	if err != nil {
		return errors.Wrap(err, "Sending data to Logstash failed")
	}

	return nil
}
