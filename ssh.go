package main

import (
	"errors"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/cloudfoundry-incubator/diego-ssh/helpers"
	"github.com/cloudfoundry/cli/plugin"
	"github.com/sykesm/cf-ssh-plugin/models/app"
	"github.com/sykesm/cf-ssh-plugin/models/credential"
	"github.com/sykesm/cf-ssh-plugin/models/info"
	"github.com/sykesm/cf-ssh-plugin/options"
)

type SshPlugin struct {
	AppFactory  app.AppFactory
	InfoFactory info.InfoFactory
	CredFactory credential.CredentialFactory
}

func (c *SshPlugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "SSH",
		Version: plugin.VersionType{
			Major: 0,
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
	plugin.Start(new(SshPlugin))
}

func (c *SshPlugin) Run(cli plugin.CliConnection, args []string) {
	c.AppFactory = app.NewAppFactory(cli)
	c.InfoFactory = info.NewInfoFactory(cli)
	c.CredFactory = credential.NewCredentialFactory(cli)

	if args[0] == "ssh" {
		opts := &options.Options{}
		err := opts.Parse(args[1:])
		if err != nil {
			fmt.Println("Invalid usage:", err)
			c.showUsage()
		}
	}
}

func (c *SshPlugin) RunWithOptions(cli plugin.CliConnection, opts *options.Options) {
	app, err := c.AppFactory.Get(opts.AppName)
	if err != nil {
		fmt.Println(err)
		return
	}

	info, err := c.InfoFactory.Get()
	if err != nil {
		fmt.Println(err)
		return
	}

	cred, err := c.CredFactory.Get()
	fmt.Println("1")
	if err != nil {
		fmt.Println("2")
		fmt.Println(err)
		return
	}
	fmt.Println("3")

	hostKeyCallback := func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		switch len(info.SSHEndpointFingerprint) {
		case 0:
			return nil
		case helpers.SHA1_FINGERPRINT_LENGTH:
			if helpers.SHA1Fingerprint(key) != info.SSHEndpointFingerprint {
				return errors.New("Host fingerprint does not match")
			}
		case helpers.MD5_FINGERPRINT_LENGTH:
			if helpers.MD5Fingerprint(key) != info.SSHEndpointFingerprint {
				return errors.New("Host fingerprint does not match")
			}
		default:
			return errors.New("invalid fingerprint format")
		}
		return nil
	}
	if opts.SkipHostValidation {
		hostKeyCallback = nil
	}

	clientConfig := &ssh.ClientConfig{
		User: fmt.Sprintf("cf:%s/%d", app.Guid, opts.Instance),
		Auth: []ssh.AuthMethod{
			ssh.Password(cred.Token),
		},
		HostKeyCallback: hostKeyCallback,
	}

	fmt.Println("daemon 1")
	_, err = ssh.Dial("tcp", info.SSHEndpoint, clientConfig)
	fmt.Println("daemon 2", err)
	if err != nil {
		fmt.Printf("FAILED\n%s\n", err.Error())
		return
	}
}

func (c *SshPlugin) showUsage() {
	fmt.Println("NAME:")
	fmt.Println("   ssh")
	fmt.Println("USAGE:")
	fmt.Println("   " + c.GetMetadata().Commands[0].UsageDetails.Usage)
}
