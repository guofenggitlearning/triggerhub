package config

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// PathEntry defines the settings of a path to back up
type PathEntry struct {
	ID    string
	Path  string
	Token string `mapstructure:"token"`
}

// Init does the initial viper setup
func Init(rootCmd *cobra.Command) {
	// config file
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("cert", rootCmd.PersistentFlags().Lookup("cert"))
	viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("tls", rootCmd.PersistentFlags().Lookup("tls"))

	viper.AddConfigPath(".")

	var configFile = viper.GetString("config")
	viper.SetConfigFile(configFile)
	// viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
