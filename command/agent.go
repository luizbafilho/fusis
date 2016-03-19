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
	"github.com/luizbafilho/fusis/fusis"
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
		agent, err := fusis.NewAgent()
		if err != nil {
			panic(err)
		}

		err = agent.Start(fusisConfig)
		if err != nil {
			panic(err)
		}
		_, err = agent.Join([]string{balancerIP}, true)
		if err != nil {
			panic(err)
		}

		waitSignals()
	},
}

var balancerIP string

func init() {
	FusisCmd.AddCommand(agentCmd)

	agentCmd.Flags().StringVarP(&balancerIP, "balancer", "b", "", "Balancer IP address.")
}
