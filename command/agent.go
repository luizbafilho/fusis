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
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var agentConfig config.AgentConfig

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Fusis Agent",
	Long: `fusis agent is the command used to run the agent process.

It's responsible for join the balancer cluster and configuring the host network
properly in orderj to enable correct IPVS balancing.`,
	Run: runAgentCmd,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&agentConfig)
	},
}

func runAgentCmd(cmd *cobra.Command, args []string) {
	agent, err := fusis.NewAgent(&agentConfig)
	if err != nil {
		panic(err)
	}

	err = agent.Start()
	if err != nil {
		panic(err)
	}
	_, err = agent.Join([]string{agentConfig.Balancer}, true)
	if err != nil {
		panic(err)
	}

	waitSignals(agent)
}

func init() {
	FusisCmd.AddCommand(agentCmd)
	setupConfig()
}

func setupConfig() {
	hostname, _ := os.Hostname()
	agentCmd.Flags().StringVarP(&agentConfig.Balancer, "balancer", "b", "", "master balancer IP address")
	agentCmd.Flags().StringVarP(&agentConfig.Name, "name", "n", hostname, "node name (unique in the cluster)")
	agentCmd.Flags().StringVar(&agentConfig.Host, "host", "", "host IP address")
	agentCmd.Flags().Uint16VarP(&agentConfig.Port, "port", "p", 80, "port number")
	agentCmd.Flags().Int32VarP(&agentConfig.Weight, "weight", "w", 1, "host weigth")
	agentCmd.Flags().StringVarP(&agentConfig.Mode, "mode", "m", "nat", "host IP address")
	agentCmd.Flags().StringVar(&agentConfig.Service, "service", "", "service id")
	agentCmd.Flags().StringVar(&agentConfig.Interface, "iface", "eth0", "Network interface")

	err := viper.BindPFlags(agentCmd.Flags())
	if err != nil {
		log.Errorf("error binding pflags: %v", err)
	}
}
