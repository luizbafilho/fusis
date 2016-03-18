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
	"log"
	"os"

	"github.com/google/seesaw/ipvs"
	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/cluster"
	"github.com/spf13/cobra"
)

// balancerCmd represents the balancer command
var balancerCmd = &cobra.Command{
	Use:   "balancer",
	Short: "Fusis Balancer",
	Long: `fusis balancer is the command used to run the balancer process.

It's responsible for creating new Services and watching for Agents joining the cluster,
and add routes to them in the Load Balancer.`,
	Run: run,
}

func init() {
	FusisCmd.AddCommand(balancerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// balancerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// balancerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func run(cmd *cobra.Command, args []string) {
	if err := ipvs.Init(); err != nil {
		log.Fatalf("IPVS initialisation failed: %v\n", err)
	}
	log.Printf("IPVS version %s\n", ipvs.Version())

	env := os.Getenv("FUSIS_ENV")
	if env == "" {
		env = "development"
	}

	balancer, err := cluster.NewBalancer()
	if err != nil {
		panic(err)
	}

	err = balancer.Start(bindAddr)
	if err != nil {
		panic(err)
	}

	apiService := api.NewAPI(env)
	apiService.Serve()
}
