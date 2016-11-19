package command

import (
	"crypto/rand"
	"fmt"
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
		Run: balancerCommandFunc,
	}

	setupDefaultOptions()
	setupBalancerCmdFlags(cmd)

	return cmd
}

func setupDefaultOptions() {
	viper.SetDefault("interfaces", map[string]string{
		"Inbound":  "eth0",
		"Outbound": "eth1",
	})

	viper.SetDefault("cluster-mode", "unicast")
	viper.SetDefault("data-path", "/etc/fusis")
	viper.SetDefault("name", randStr())
	viper.SetDefault("log-level", "warn")
}

func setupBalancerCmdFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&conf.Bootstrap, "bootstrap", false, "Starts balancer in boostrap mode")
	cmd.Flags().BoolVar(&conf.DevMode, "dev", false, "Initialize balancer in dev mode")
	cmd.Flags().StringSliceVarP(&conf.Join, "join", "j", []string{}, "Join balancer pool")
	cmd.Flags().StringVar(&configFile, "config", "", "specify a configuration file")

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		log.Errorf("error binding pflags: %v", err)
	}
}

func balancerCommandFunc(cmd *cobra.Command, args []string) {
	if err := net.SetIpForwarding(); err != nil {
		log.Warn("Fusis couldn't set net.ipv4.ip_forward=1")
		log.Fatal(err)
	}

	err := conf.Validate()
	if err != nil {
		fmt.Println("Error: Invalid configuration file.", err)
		os.Exit(1)
	}

	balancer, err := fusis.NewBalancer(&conf)
	if err != nil {
		log.Fatal(err)
	}

	apiService := api.NewAPI(balancer)
	go apiService.Serve()

	waitSignals(balancer)
}

func randStr() string {
	dictionary := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, 15)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}
