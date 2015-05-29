package main

import (
	"fmt"

	"github.com/cloudfoundry/cli/plugin"
	"github.com/sykesm/cf-ssh-plugin/options"
)

type SshPlugin struct{}

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

}

func (c *SshPlugin) showUsage() {
	fmt.Println("NAME:")
	fmt.Println("   ssh")
	fmt.Println("USAGE:")
	fmt.Println("   " + c.GetMetadata().Commands[0].UsageDetails.Usage)
}
