package command

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var FusisCmd = &cobra.Command{
	Use:   "fusis",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := FusisCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	customFormatter := new(log.TextFormatter)
  customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(configFile)
	}

	viper.SetConfigName("fusis") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")            // adding current directory as first search path
	viper.AddConfigPath("$HOME/.fusis") // adding ~/.fusis as fallback
	viper.AddConfigPath("/etc")         // adding /etc as next fallback
	viper.AutomaticEnv()                // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config from %s", viper.ConfigFileUsed())
	} else {
		log.Fatal("Error reading config file: fusis.toml", err)
		log.Fatal("Searched: ./, ~/fusis and /etc")
	}

}

type Node interface {
	Shutdown()
}

func waitSignals(node Node) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

	node.Shutdown()
}
