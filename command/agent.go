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
	"fmt"

	"github.com/hashicorp/serf/serf"
	"github.com/luizbafilho/fusis/cluster"
	"github.com/spf13/cobra"
)

// agentCmd represents the balancer command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Fusis Agent",
	Long: `fusis agent is the command used to run the agent process.

It's responsible for join the balancer cluster and configuring the host network
properly in order to enable correct IPVS balancing.`,
	Run: func(cmd *cobra.Command, args []string) {
		agent, err := cluster.NewAgent()
		if err != nil {
			panic(err)
		}

		err = agent.Start(bindAddr)
		if err != nil {
			panic(err)
		}
		_, err = agent.Join([]string{balancerIP}, true)
		if err != nil {
			panic(err)
		}

		eventCh := make(chan serf.Event, 64)
		for {
			select {
			case e := <-eventCh:
				fmt.Printf("[INFO] fusis agent: Received event: %s", e.String())
			}
		}
	},
}

var balancerIP string

func init() {
	FusisCmd.AddCommand(agentCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// agentCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// agentCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	agentCmd.Flags().StringVarP(&balancerIP, "balancer", "b", "", "Balancer IP address.")
}
