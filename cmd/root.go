package cmd

import (
	"fmt"
	"log"
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
		Run:   mainHandler,
	}
)

// Execute runs the cobra commands and parameters
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// MAIN ENTRY POINT
func mainHandler(cmd *cobra.Command, args []string) {
	var configFile = viper.GetString("config")
	var port = viper.GetInt("port")
	var authToken = viper.GetString("token")
	var useTLS = viper.GetBool("tls")
	var cert = viper.GetString("cert")
	var key = viper.GetString("key")

	if configFile != "" {
		log.Output(1, fmt.Sprintf("Using config file %s", configFile))
	}

	server.Run(port, authToken, useTLS, cert, key)
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
