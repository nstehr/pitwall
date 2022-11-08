package core

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func SSH(port int) error {
	ssh.Handle(func(s ssh.Session) {
		cmd := exec.Command("/bin/sh")
		ptyReq, winCh, isPty := s.Pty()
		if isPty {
			cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
			f, err := pty.Start(cmd)
			if err != nil {
				log.Println(err)
			}
			go func() {
				for win := range winCh {
					setWinsize(f, win.Width, win.Height)
				}
			}()
			go func() {
				io.Copy(f, s) // stdin
			}()
			io.Copy(s, f) // stdout
			cmd.Wait()
		} else {
			io.WriteString(s, "No PTY requested.\n")
			s.Exit(1)
		}
	})

	log.Printf("starting ssh server on port %d...\n", port)
	publicKeyOption := ssh.PublicKeyAuth(func(ctx ssh.Context, key ssh.PublicKey) bool {
		f, err := os.ReadFile("/etc/powerunit/key")
		if err != nil {
			log.Println("Error reading public key for SSH", err)
			return false
		}
		allowed, _, _, _, err := ssh.ParseAuthorizedKey(f)
		if err != nil {
			log.Println("Error parsing public key", err)
			return false
		}

		return ssh.KeysEqual(key, allowed)
	})
	return ssh.ListenAndServe(fmt.Sprintf(":%d", port), nil, publicKeyOption)
}
