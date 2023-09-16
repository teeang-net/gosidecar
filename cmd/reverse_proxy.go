package cmd

import (
	"github.com/davidtaing/gosidecar/internal/reverse_proxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var target string
var port uint32

var reverseProxyCmd = &cobra.Command{
	Use:   "reverseproxy",
	Short: "Starts a reverse proxy",
	Run: func(cmd *cobra.Command, args []string) {
		reverse_proxy.Start(target, port)
	},
}

func init() {
	rootCmd.AddCommand(reverseProxyCmd)

	reverseProxyCmd.Flags().StringVarP(&target, "target", "t", "", "The target URL to proxy to")
	reverseProxyCmd.Flags().Uint32VarP(&port, "port", "p", 8080, "The port to listen on")
	reverseProxyCmd.MarkFlagRequired("target")
	viper.SetDefault("port", 8080)
}
