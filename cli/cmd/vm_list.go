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

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all of the user's virtual machines",

	Run: func(cmd *cobra.Command, args []string) {
		cfg := cmd.Context().Value(contextConfig).(*config.Config)
		virtualMachines, err := api.GetVMs(cmd.Context(), cfg)
		if err != nil {
			log.Fatal("Error getting VMs", err)
		}
		val, err := virtualMachines.PrettyString()
		if err != nil {
			log.Fatal("Error getting VMs", err)
		}
		log.Println(val)
	},
}

func init() {
	vmCmd.AddCommand(listCmd)
}
