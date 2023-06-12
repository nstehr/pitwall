/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/nstehr/pitwall/cli/internal/config"
	"github.com/spf13/cobra"
)

const (
	authEndpointFlag = "authEndpoint"
	apiEndpointFlag  = "apiEndpoint"
)

var (
	authEndpoint string
	apiEndpoint  string
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure pwctl to be able to authenticate and run commands",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := cmd.Context().Value(contextConfig).(*config.Config)
		cfg.ApiEndpoint = apiEndpoint
		cfg.AuthEndpoint = authEndpoint
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.Flags().StringVar(&authEndpoint, authEndpointFlag, "", "Endpoint to used to retrieve authentication token")
	configCmd.Flags().StringVar(&apiEndpoint, apiEndpointFlag, "", "Endpoint to used to perform API requests against")
}
