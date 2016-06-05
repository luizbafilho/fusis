package command

import (
	log "github.com/Sirupsen/logrus"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/luizbafilho/fusis/net"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var balancerCmd = &cobra.Command{
	Use:   "balancer",
	Short: "Fusis Balancer",
	Long: `fusis balancer is the command used to run the balancer process.

It's responsible for creating new Services and watching for Agents joining the cluster,
and add routes to them in the Load Balancer.`,
	Run: run,
	PreRun: func(cmd *cobra.Command, args []string) {
		viper.Unmarshal(&config.Balancer)
	},
}

func init() {
	FusisCmd.AddCommand(balancerCmd)
	setupBalancerConfig()
}

func setupBalancerConfig() {
	balancerCmd.Flags().StringVarP(&config.Balancer.Interface, "interface", "", "eth0", "Network interface")
	balancerCmd.Flags().StringVarP(&config.Balancer.Join, "join", "j", "", "Join balancer pool")
	balancerCmd.Flags().BoolVarP(&config.Balancer.Single, "single", "s", false, "Configuration directory")
	balancerCmd.Flags().StringVarP(&config.Balancer.ConfigPath, "config-path", "", "/etc/fusis", "Configuration directory")
	balancerCmd.Flags().IntVar(&config.Balancer.RaftPort, "raft-port", 4382, "Raft port")

	err := viper.BindPFlags(balancerCmd.Flags())
	if err != nil {
		log.Errorf("error binding pflags: %v", err)
	}
}

func run(cmd *cobra.Command, args []string) {
	if err := net.SetIpForwarding(); err != nil {
		log.Warn("Fusis couldn't set net.ipv4.ip_forward=1")
		log.Fatal(err)
	}

	balancer, err := fusis.NewBalancer()
	if err != nil {
		panic(err)
	}

	if config.Balancer.Join != "" {
		balancer.JoinPool()
	}

	apiService := api.NewAPI(balancer)
	go apiService.Serve()

	waitSignals(balancer)
}
