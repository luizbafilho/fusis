package command

import (
	"crypto/rand"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/luizbafilho/fusis/api"
	"github.com/luizbafilho/fusis/config"
	"github.com/luizbafilho/fusis/fusis"
	"github.com/luizbafilho/fusis/net"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	conf       config.BalancerConfig
	configFile string
)

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

	level, _ := log.ParseLevel(strings.ToUpper(conf.LogLevel))
	log.Info(log.WarnLevel, level)
	log.SetLevel(log.DebugLevel)

	return cmd
}

func setupDefaultOptions() {
	viper.SetDefault("cluster-mode", "unicast")
	viper.SetDefault("data-path", "/etc/fusis")
	viper.SetDefault("name", randStr())
	viper.SetDefault("log-level", "warn")
	viper.SetDefault("enable-health-checks", true)
	viper.SetDefault("store-prefix", "/fusis")
}

func setupBalancerCmdFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&configFile, "config", "", "specify a configuration file")
	cmd.Flags().StringVar(&conf.LogLevel, "log-level", "", "specify a log level")
	cmd.Flags().BoolVarP(&conf.EnableHealthChecks, "enable-health-checks", "", true, "enables health checking on destinations")
	cmd.Flags().StringVarP(&conf.StorePrefix, "store-prefix", "", "fusis", "configures the prefix used by the store")
	// cmd.Flags().StringVarP(&conf.StoreAddress, "store-address", "", "consul://localhost:8500", "configures the store address")

	err := viper.BindPFlags(cmd.Flags())
	if err != nil {
		log.Errorf("Error binding pflags: %v", err)
	}
}

func balancerCommandFunc(cmd *cobra.Command, args []string) {
	if err := net.SetIpForwarding(); err != nil {
		log.Warn("Fusis couldn't set ip forwarding in the kernel with net.ipv4.ip_forward=1")
		log.Fatal(err)
	}

	if err := conf.Validate(); err != nil {
		log.Fatal("Invalid configuration file: ", err)
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
