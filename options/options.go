package options

import (
	"errors"

	"github.com/cloudfoundry/cli/flags"
	"github.com/cloudfoundry/cli/flags/flag"
)

type Options struct {
	AppName            string
	Instance           int
	LocalPort          uint16
	ForwardDestination string
	LocalProxy         bool
	LocalPortForward   bool
}

var UsageError = errors.New("Invalid usage")

func (o *Options) Parse(args []string) error {
	if len(args) == 0 {
		return UsageError
	}

	fc := flags.NewFlagContext(setupFlags())
	err := fc.Parse(args...)
	if err != nil {
		return err
	}

	if len(fc.Args()) != 1 {
		return UsageError
	}

	o.AppName = fc.Args()[0]

	if fc.IsSet("i") {
		instance := fc.Int("i")
		if instance < 0 {
			return errors.New("Value for flag 'i' must not be negative")
		}

		o.Instance = fc.Int("i")
	}

	return nil
}

func setupFlags() map[string]flags.FlagSet {
	fs := make(map[string]flags.FlagSet)
	fs["i"] = &cliFlags.IntFlag{Name: "i", Usage: ""}
	return fs
}
