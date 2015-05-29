package main_test

import (
	"github.com/cloudfoundry/cli/plugin/fakes"
	"github.com/sykesm/cf-ssh-plugin"
	"github.com/sykesm/cf-ssh-plugin/app/app_fakes"

	io_helpers "github.com/cloudfoundry/cli/testhelpers/io"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("DiegoSsh", func() {
	var (
		fakeCliConnection    *fakes.FakeCliConnection
		callCliCommandPlugin *main.SshPlugin
		fakeAppFactory       *app_fakes.FakeAppFactory
	)

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
		callCliCommandPlugin = &main.SshPlugin{}
		fakeAppFactory = &app_fakes.FakeAppFactory{}
	})

	Describe("command arguments", func() {
		Context("when arguments are invalid", func() {
			It("presents a help message", func() {
				output := io_helpers.CaptureOutput(func() {
					callCliCommandPlugin.Run(fakeCliConnection, []string{"ssh"})
				})

				Expect(output).To(ContainSubstrings(
					[]string{"Invalid usage"},
					[]string{"NAME:"},
					[]string{"USAGE:"},
				))
			})
		})
	})

	Describe("Getting app model", func() {
		Context("when there is an error getting the app model", func() {
		})

		Context("when the app model is valid", func() {
		})

	})
})
