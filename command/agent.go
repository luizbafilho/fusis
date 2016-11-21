package command

import (
	log "github.com/Sirupsen/logrus"

	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	agentConfig config.AgentConfig
	configFile  string
)

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
		log.Fatal(err)
	}

	err = agent.Start()
	if err != nil {
		log.Fatal(err)
	}
	_, err = agent.Join([]string{agentConfig.Balancer}, true)
	if err != nil {
		log.Fatal(err)
	}

	waitSignals(agent)
}

func init() {
	FusisCmd.AddCommand(agentCmd)
	setupDefaultAgentOptions()
	setupConfig()
}

func setupDefaultAgentOptions() {
	viper.SetDefault("name", randStr())
	viper.SetDefault("address", "")
	viper.SetDefault("weight", 1)
	viper.SetDefault("interface", "eth0")
}

func setupConfig() {
	agentCmd.Flags().StringVar(&configFile, "config", "", "specify a configuration file")

	err := viper.BindPFlags(agentCmd.Flags())
	if err != nil {
		log.Errorf("Error binding pflags: %v", err)
	}
}
