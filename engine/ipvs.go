// Copyright 2012 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: jsing@google.com (Joel Sing)

package engine

import (
	"sync"

	log "github.com/golang/glog"
	"github.com/google/seesaw/ipvs"
)

var mt sync.Mutex

// initIPVS initialises the IPVS sub-component.
func initIPVS() {
	mt.Lock()
	defer mt.Unlock()
	log.Infof("Initialising IPVS...")
	if err := ipvs.Init(); err != nil {
		// TODO(jsing): modprobe ip_vs and try again.
		log.Fatalf("IPVS initialisation failed: %v", err)
	}
	log.Infof("IPVS version %s", ipvs.Version())
}

// IPVSFlush flushes all services and destinations from the IPVS table.
func IPVSFlush() error {
	mt.Lock()
	defer mt.Unlock()
	return ipvs.Flush()
}

// IPVSGetServices gets the currently configured services from the IPVS table.
func IPVSGetServices() ([]*ipvs.Service, error) {
	mt.Lock()
	defer mt.Unlock()
	return ipvs.GetServices()
}

// IPVSGetService gets the currently configured service from the IPVS table,
// which matches the specified service.
// func IPVSGetService(svc ipvs.Service) (error) {
// 	mt.Lock()
// 	defer mt.Unlock()
// 	so, err := ipvs.GetService(svc)
// 	if err != nil {
// 		return err
// 	}
// 	s.Services = []*ipvs.Service{so}
// 	return nil
// }
//
// // IPVSAddService adds the specified service to the IPVS table.
func IPVSAddService(svc *ipvs.Service) error {
	mt.Lock()
	defer mt.Unlock()
	return ipvs.AddService(*svc)
}

//
// // IPVSUpdateService updates the specified service in the IPVS table.
// func IPVSUpdateService(svc *ipvs.Service, out *int) error {
// 	mt.Lock()
// 	defer mt.Unlock()
// 	return ipvs.UpdateService(*svc)
// }
//
// // IPVSDeleteService deletes the specified service from the IPVS table.
// func IPVSDeleteService(svc *ipvs.Service, out *int) error {
// 	mt.Lock()
// 	defer mt.Unlock()
// 	return ipvs.DeleteService(*svc)
// }
//
// IPVSAddDestination adds the specified destination to the IPVS table.
func IPVSAddDestination(svc ipvs.Service, dst ipvs.Destination) error {
	mt.Lock()
	defer mt.Unlock()
	return ipvs.AddDestination(svc, dst)
}

//
// // IPVSUpdateDestination updates the specified destination in the IPVS table.
// func IPVSUpdateDestination(dst *ncctypes.IPVSDestination, out *int) error {
// 	mt.Lock()
// 	defer mt.Unlock()
// 	return ipvs.UpdateDestination(*dst.Service, *dst.Destination)
// }
//
// // IPVSDeleteDestination deletes the specified destination from the IPVS table.
func IPVSDeleteDestination(svc ipvs.Service, dst ipvs.Destination) error {
	mt.Lock()
	defer mt.Unlock()
	return ipvs.DeleteDestination(svc, dst)
}
