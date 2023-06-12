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
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get's the given virtual machine",
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := cmd.Context().Value(contextConfig).(*config.Config)
		virtualMachine, err := api.GetVMByName(cmd.Context(), cfg, args[0])
		if err != nil {
			log.Fatal("Error getting VM", err)
		}
		val, err := virtualMachine.PrettyString()
		if err != nil {
			log.Fatal("Error getting VM", err)
		}
		log.Println(val)
	},
}

func init() {
	vmCmd.AddCommand(getCmd)
}
