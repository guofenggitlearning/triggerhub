package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// DispatcherEntry contains the settings of the remote dispatcher
type DispatcherEntry struct {
	URL   string `mapstructure:"url"`
	Token string `mapstructure:"token"`
}

// ActionEntry defines an action supported by the listener
type ActionEntry struct {
	Name    string `mapstructure:"name"`
	Token   string `mapstructure:"token"`
	Command string `mapstructure:"command"`
}

// DispatcherInit does the initial viper setup
func DispatcherInit(rootCmd *cobra.Command, serveCmd *cobra.Command) {
	// config file
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("port", serveCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("cert", serveCmd.PersistentFlags().Lookup("cert"))
	viper.BindPFlag("key", serveCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("tls", serveCmd.PersistentFlags().Lookup("tls"))

	viper.AddConfigPath(".")

	var configFile = viper.GetString("config")
	viper.SetConfigFile(configFile)
	// viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// ListenerInit does the initial viper setup
func ListenerInit(rootCmd *cobra.Command) {
	// config file
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	viper.AddConfigPath(".")

	var configFile = viper.GetString("config")
	if configFile == "" {
		return
	}

	viper.SetConfigFile(configFile)

	if err := viper.ReadInConfig(); err == nil {
		// fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
