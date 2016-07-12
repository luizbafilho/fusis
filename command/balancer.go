package command

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/luizbafilho/fusis/net"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var conf config.BalancerConfig

func init() {
	FusisCmd.AddCommand(NewBalancerCommand())
}

func NewBalancerCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balancer [options]",
		Short: "starts a new balancer",
		Long: `fusis balancer is the command used to run the balancer process.

	It's responsible for creating new Services and watching for Agents joining the cluster,
	and add routes to them in the Load Balancer.`,
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.Unmarshal(&conf)
		},
		RunE: balancerCommandFunc,
	}

	setupBalancerCmdFlags(cmd)

	return cmd
}

func setupBalancerCmdFlags(cmd *cobra.Command) {
	hostname, _ := os.Hostname()
	cmd.Flags().StringVarP(&conf.Name, "name", "n", hostname, "node name (unique in the cluster)")
	cmd.Flags().StringVarP(&conf.Interface, "interface", "", "eth0", "Network interface")
	cmd.Flags().StringVarP(&conf.ConfigPath, "config-path", "", "/etc/fusis", "Configuration directory")
	cmd.Flags().BoolVar(&conf.Bootstrap, "bootstrap", false, "starts balancer in boostrap mode")
	cmd.Flags().BoolVar(&conf.DevMode, "dev", false, "Initialize balancer in dev mode")
	cmd.Flags().StringSliceVarP(&conf.Join, "join", "j", []string{}, "Join balancer pool")
	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		log.Errorf("error binding pflags: %v", err)
	}
}

func balancerCommandFunc(cmd *cobra.Command, args []string) error {
	if err := net.SetIpForwarding(); err != nil {
		log.Warn("Fusis couldn't set net.ipv4.ip_forward=1")
		log.Fatal(err)
	}

	balancer, err := fusis.NewBalancer(&conf)
	if err != nil {
		log.Fatal(err)
	}

	if len(conf.Join) > 0 {
		balancer.JoinPool()
	}

	apiService := api.NewAPI(balancer)
	go apiService.Serve()

	waitSignals(balancer)

	return nil
}
