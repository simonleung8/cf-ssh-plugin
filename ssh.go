package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/cloudfoundry-incubator/diego-ssh/sigwinch"

	"code.google.com/p/go.crypto/ssh"
	"github.com/docker/docker/pkg/term"

	"github.com/cloudfoundry/cli/plugin"
)

type SshCmd struct{}

func (c *SshCmd) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "SSH",
		Version: plugin.VersionType{
			Major: 1,
			Minor: 1,
			Build: 0,
		},
		Commands: []plugin.Command{
			{
				Name:     "ssh",
				HelpText: "ssh to an application container instance",
				UsageDetails: plugin.Usage{
					Usage: "cf ssh APP-NAME [-i instance]",
				},
			},
		},
	}
}

func main() {
	plugin.Start(new(SshCmd))
}

func (c *SshCmd) Run(cli plugin.CliConnection, args []string) {
	index := 0
	if len(args) > 2 && flagPresent(args, "-i") {
		idxFlag := getFlagValue(args, "-i")

		var err error
		index, err = strconv.Atoi(idxFlag)
		if err != nil {
			fmt.Printf("Invalid index: %q\n", idxFlag)
			return
		}
	}

	result, err := cli.CliCommandWithoutTerminalOutput("app", args[1], "--guid")
	if err != nil {
		if strings.Contains(result[0], "FAILED") {
			fmt.Printf("Application not found\n")
			return
		}
	}

	guid := strings.TrimSpace(result[0])

	result, err = cli.CliCommandWithoutTerminalOutput("oauth-token")
	if err != nil {
		if strings.Contains(result[1], "FAILED") {
			fmt.Printf("Failed to acquire authorization token\n")
			return
		}
	}

	token := result[3]

	config := &ssh.ClientConfig{
		User: fmt.Sprintf("cf:%s/%d", guid, index),
		Auth: []ssh.AuthMethod{
			ssh.Password(token),
		},
	}

	client, err := ssh.Dial("tcp", "ssh.ketchup.cf-app.com:2222", config)
	if err != nil {
		fmt.Printf("SSH authentication failed\n")
		return
	}
	defer client.Close()

	c.interactiveSession(client)
}

func (c *SshCmd) interactiveSession(client *ssh.Client) {
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to allocate SSH session\n")
		return
	}
	defer session.Close()

	stdin, stdout, stderr := term.StdStreams()
	session.Stdout = stdout
	session.Stderr = stderr
	in, err := session.StdinPipe()
	if err != nil {
		fmt.Printf("Failed to connect stdin\n")
		return
	}

	stdinFd, _ := term.GetFdInfo(stdin)
	stdoutFd, _ := term.GetFdInfo(stdout)

	state, err := term.SetRawTerminal(stdinFd)
	if err != nil {
		fmt.Printf("Failed to make stdin raw\n")
		return
	}
	defer term.RestoreTerminal(stdinFd, state)

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 115200,
		ssh.TTY_OP_OSPEED: 115200,
	}

	width, height, err := getWindowDimentions(stdoutFd)
	if err != nil {
		width = 80
		height = 43
	}

	err = session.RequestPty("xterm", height, width, modes)
	if err != nil {
		fmt.Printf("PTY allocation failed\n")
		return
	}

	err = session.Shell()
	if err != nil {
		fmt.Printf("Shell creation failed\n")
		return
	}

	go func() {
		io.Copy(in, stdin)
	}()

	resized := make(chan os.Signal, 16)
	if runtime.GOOS == "windows" {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		go func() {
			for _ = range ticker.C {
				resized <- syscall.Signal(-1)
			}
		}()
	} else {
		signal.Notify(resized, sigwinch.SIGWINCH())
	}

	go resize(resized, session, stdoutFd)

	session.Wait()
}

func resize(resized <-chan os.Signal, session *ssh.Session, terminalFd uintptr) {
	type resizeMessage struct {
		Width       uint32
		Height      uint32
		PixelWidth  uint32
		PixelHeight uint32
	}

	previousWidth, previousHeight, _ := getWindowDimentions(terminalFd)

	for _ = range resized {
		width, height, err := getWindowDimentions(terminalFd)
		if err != nil {
			continue
		}

		if width == previousWidth && height == previousHeight {
			continue
		}

		message := resizeMessage{
			Width:  uint32(width),
			Height: uint32(height),
		}

		session.SendRequest("window-change", false, ssh.Marshal(message))

		previousWidth = width
		previousHeight = height
	}
}

func getWindowDimentions(terminalFd uintptr) (width int, height int, err error) {
	winSize, err := term.GetWinsize(terminalFd)
	if err != nil {
		return 0, 0, err
	}

	return int(winSize.Width), int(winSize.Height), nil
}

func flagPresent(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func getFlagValue(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag {
			if len(args) >= i+1 {
				return args[i+1]
			}
			break
		}
	}
	return ""
}
