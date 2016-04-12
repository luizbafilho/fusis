// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
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

package command

import (
	log "github.com/Sirupsen/logrus"

	"os"

	"github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/luizbafilho/fusis/net"
	"github.com/spf13/cobra"
)

var balancerCmd = &cobra.Command{
	Use:   "balancer",
	Short: "Fusis Balancer",
	Long: `fusis balancer is the command used to run the balancer process.

It's responsible for creating new Services and watching for Agents joining the cluster,
and add routes to them in the Load Balancer.`,
	Run: run,
}

func init() {
	balancerCmd.Flags().StringVarP(&fusisConfig.Interface, "iface", "", "eth0", "Network interface")
	FusisCmd.AddCommand(balancerCmd)
}

func run(cmd *cobra.Command, args []string) {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v", err)
	}
	log.Printf("IPVS version %s", ipvs.Version())

	if err := net.SetIpForwarding(); err != nil {
		log.Fatal(err)
		log.Warn("Fusis couldn't set net.ipv4.ip_forward=1")
	}

	env := os.Getenv("FUSIS_ENV")
	if env == "" {
		env = "development"
	}

	balancer, err := fusis.NewBalancer()
	if err != nil {
		panic(err)
	}

	err = balancer.Start(fusisConfig)
	if err != nil {
		panic(err)
	}

	apiService := api.NewAPI(env)
	go apiService.Serve()

	waitSignals()
}
