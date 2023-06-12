/*
Copyright © 2023 Nathan Stehr nstehr@gmail.com
*/
package cmd

import (
	"context"
	"log"
	"os"

	"github.com/nstehr/pitwall/cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
)

type contextConfigKey string
type contextOAuthTokenKey string

const (
	contextConfig     = contextConfigKey("cfg")
	contextOAuthToken = contextOAuthTokenKey("oauthToken")
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "pwctl",
	Short: "CLI for interacting with pitwall API",
	Long:  "CLI for interacting with pitwall API",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		exists := config.ConfigExists(cfgFile)
		if !exists {
			if cmd.CalledAs() != "config" {
				log.Fatalf("%s does not exist.  Please run pwctl config to initialize\n", cfgFile)
			}
		}
		// a bit of a hack, but we detect if it is the config command and then will get or create the file
		// here, so all the commands will just access the config via the context
		cfg, err := config.GetOrCreateConfig(cfgFile)
		if err != nil {
			log.Fatal("Error with config file: ", err)
		}
		ctx := context.WithValue(cmd.Context(), contextConfig, cfg)
		cmd.SetContext(ctx)
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		f := cmd.Context().Value(contextConfig).(*config.Config)
		err := f.WriteToFile(cfgFile)
		if err != nil {
			log.Fatal("Could not save config file")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", ".pitwall.json", "")
}
