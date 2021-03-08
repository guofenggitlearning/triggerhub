package cmd

import (
	"fmt"
	"os"

	"github.com/brickpop/packerd/config"
	"github.com/brickpop/packerd/server"
	"github.com/spf13/cobra"
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
	cobra.OnInitialize(func() {
		config.Init(rootCmd)
	})

	// Read flags
	rootCmd.PersistentFlags().String("config", "", "the config file to use")
	rootCmd.PersistentFlags().String("cert", "", "the certificate file (TLS only)")
	rootCmd.PersistentFlags().String("key", "", "the TLS encryption key file")
	rootCmd.PersistentFlags().Bool("tls", false, "whether to use TLS encryption (cert and key required)")
	rootCmd.PersistentFlags().IntP("port", "p", 8080, "port to bind to")
}
