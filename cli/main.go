package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/openziti/sdk-golang/ziti"
	"github.com/openziti/sdk-golang/ziti/config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

var (
	identity = flag.String("identity", "", "Ziti Identity file")
	service  = flag.String("service", "", "Ziti Service")
)

func main() {
	flag.Parse()
	cfg, err := config.NewFromFile(*identity)
	if err != nil {
		log.Fatalf("failed to load config err=%v", err)
	}

	ztx := ziti.NewContextWithConfig(cfg)
	err = ztx.Authenticate()
	log.Println(err)
	conn, err := ztx.Dial(*service)
	log.Println(err)

	pKey, err := os.ReadFile("id_bob")
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
	//session.Wait()

	// Clean exit.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig
	log.Println("Ctrl-c detected, shutting down")
	log.Println("Goodbye.")

}
