/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// vmCmd represents the vm command
var vmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Operations for managing virtual machines",
}

func init() {
	rootCmd.AddCommand(vmCmd)
}
