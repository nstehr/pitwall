/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nstehr/pitwall/cli/internal/api"
	"github.com/nstehr/pitwall/cli/internal/config"
	"github.com/openziti/sdk-golang/ziti"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	sshPrivKey string
)

// sshCmd represents the ssh command
var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Creates an ssh connection the given VM",
	Run: func(cmd *cobra.Command, args []string) {

		// args will contain the ssh key and the name of the VM to connect to
		// the ziti identity will come from the config

		cfg := cmd.Context().Value(contextConfig).(*config.Config)
		resp, err := api.GetVMByName(cmd.Context(), cfg, args[0])
		if err != nil {
			log.Fatal("Error getting VM", err)
		}
		virtualMachine, err := resp.Parse()
		if err != nil {
			log.Fatal("Error getting VM", err)
		}
		var sshService api.Service

		for _, service := range virtualMachine.Services {
			if service.Protocol == "ssh" {
				sshService = service
				break
			}
		}
		if sshService.Name == "" {
			log.Fatal("No ssh service found")
		}
		ztx, err := ziti.NewContext(cfg.ZitiConfig)
		if err != nil {
			log.Fatal("Error getting context", err)
		}
		err = ztx.Authenticate()
		if err != nil {
			log.Fatal("Couldn't authenticate with ziti:", err)
		}
		conn, err := ztx.Dial(sshService.Name)
		if err != nil {
			log.Fatal("Couldn't dial service:", err)
		}
		pKey, err := os.ReadFile(sshPrivKey)
		if err != nil {
			log.Fatal("Error reading private key: ", err)
		}

		var signer ssh.Signer

		signer, err = ssh.ParsePrivateKey(pKey)
		if err != nil {
			fmt.Println(err.Error())
		}

		config := &ssh.ClientConfig{
			User:            "bob",
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		}
		c, chans, reqs, err := ssh.NewClientConn(conn, "", config)
		if err != nil {
			log.Fatal("Error getting ssh conn: ", err)
		}

		client := ssh.NewClient(c, chans, reqs)

		session, err := client.NewSession()
		if err != nil {
			log.Fatal("Error with ssh client session: ", err)
		}

		stdInFd := int(os.Stdin.Fd())
		stdOutFd := int(os.Stdout.Fd())

		oldState, err := term.MakeRaw(stdInFd)
		if err != nil {
			log.Fatal("Error with ssh client session: ", err)
		}
		defer func() {
			_ = session.Close()
			_ = term.Restore(stdInFd, oldState)
		}()

		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		session.Stdin = os.Stdin

		termWidth, termHeight, err := term.GetSize(stdOutFd)
		if err != nil {
			log.Fatal("Error with terminal sizing: ", err)
		}

		fmt.Println("connected.")

		if err := session.RequestPty("xterm", termHeight, termWidth, ssh.TerminalModes{ssh.ECHO: 1}); err != nil {
			log.Fatal("Error with ssh pty: ", err)
		}

		err = session.Shell()
		if err != nil {
			log.Fatal("Error with ssh shell: ", err)
		}
		session.Wait()
		_ = session.Close()
		_ = term.Restore(stdInFd, oldState)

		// Clean exit.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		<-sig
		log.Println("Ctrl-c detected, shutting down SSH connection")
		log.Println("Goodbye.")
	},
}

func init() {
	sshCmd.PersistentFlags().StringVar(&sshPrivKey, "privKey", "id_pitwall", "Private key for making connection to the pitwall VMs")
	rootCmd.AddCommand(sshCmd)
}
