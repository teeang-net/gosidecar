package cmd

import (
	"github.com/davidtaing/gosidecar/internal/rate_limiter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rateLimitTarget string
var rateLimitPort uint32

var rateLimitCmd = &cobra.Command{
	Use:   "ratelimit",
	Short: "Starts a rate limiter",
	Run: func(cmd *cobra.Command, args []string) {
		rate_limiter.Start()
	},
}

func init() {
	rootCmd.AddCommand(rateLimitCmd)

	rateLimitCmd.Flags().StringVarP(&target, "target", "t", "", "The target URL to proxy to")
	rateLimitCmd.Flags().Uint32VarP(&port, "port", "p", 8080, "The port to listen on")
	rateLimitCmd.MarkFlagRequired("target")
	viper.SetDefault("port", 8080)
}
