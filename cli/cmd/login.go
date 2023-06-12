/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/nstehr/pitwall/cli/internal/api"
	"github.com/nstehr/pitwall/cli/internal/config"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Creates a login session to the pitwall API.",
	Long:  "Creates a login session to the pitwall API.",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := cmd.Context().Value(contextConfig).(*config.Config)
		token, err := api.Authenticate(cmd.Context(), cfg.AuthEndpoint)
		if err != nil {
			log.Fatal("Error authenticating, ", err)
		}
		cfg.OAuth.RefreshToken = token.RefreshToken
		cfg.OAuth.AccessToken = token.AccessToken
		cfg.OAuth.Expiry = token.Expiry
		cfg.OAuth.TokenType = token.TokenType

		// TODO: should I move this out of the login?
		if cfg.ZitiConfig == nil {
			f, err := api.EnrollIdentity(cmd.Context(), cfg)
			if err != nil {
				log.Fatal("Error authenticating with overlay, ", err)
			}

			cfg.ZitiConfig = f
		}

	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
