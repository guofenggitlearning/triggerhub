package cmd

import (
	"fmt"
	"os"

	"github.com/brickpop/packerd/rand"
	"github.com/brickpop/packerd/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "packerd",
		Short: "Packerd is a backup utility for authenticated remote clients",
		Long:  `Packerd is a backup utility for authenticated remote clients`,
		Run: func(cmd *cobra.Command, args []string) {
			server.Run()
		},
	}
)

// Execute runs the cobra commands and parameters
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("config", "", "the config file to use")
	rootCmd.PersistentFlags().String("token", "", "the auth token to use for clients")
	rootCmd.PersistentFlags().String("cert", "", "the certificate file (TLS only)")
	rootCmd.PersistentFlags().String("key", "", "the TLS encryption key file")
	rootCmd.PersistentFlags().Bool("tls", false, "whether to use TLS encryption")
	rootCmd.PersistentFlags().IntP("port", "p", 80, "port to bind to")

	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("cert", rootCmd.PersistentFlags().Lookup("cert"))
	viper.BindPFlag("key", rootCmd.PersistentFlags().Lookup("key"))
	viper.BindPFlag("tls", rootCmd.PersistentFlags().Lookup("tls"))

	// viper.SetDefault("author", "NAME HERE <EMAIL ADDRESS>")
	viper.AddConfigPath(".")
}

func initConfig() {
	var configFile = viper.GetString("config")

	if viper.GetString("token") == "" {
		viper.Set("token", rand.String(40))
	}

	viper.SetConfigFile(configFile)
	// viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
