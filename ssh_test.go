package main_test

import (
	"errors"

	"github.com/cloudfoundry/cli/plugin/fakes"
	"github.com/sykesm/cf-ssh-plugin"
	"github.com/sykesm/cf-ssh-plugin/models/app"
	"github.com/sykesm/cf-ssh-plugin/models/app/app_fakes"
	"github.com/sykesm/cf-ssh-plugin/models/credential"
	"github.com/sykesm/cf-ssh-plugin/models/credential/credential_fakes"
	"github.com/sykesm/cf-ssh-plugin/models/info"
	"github.com/sykesm/cf-ssh-plugin/models/info/info_fakes"
	"github.com/sykesm/cf-ssh-plugin/options"

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
		fakeInfoFactory      *info_fakes.FakeInfoFactory
		fakeCredFactory      *credential_fakes.FakeCredentialFactory
	)

	BeforeEach(func() {
		fakeCliConnection = &fakes.FakeCliConnection{}
		fakeAppFactory = &app_fakes.FakeAppFactory{}
		fakeInfoFactory = &info_fakes.FakeInfoFactory{}
		fakeCredFactory = &credential_fakes.FakeCredentialFactory{}

		callCliCommandPlugin = &main.SshPlugin{
			AppFactory:  fakeAppFactory,
			InfoFactory: fakeInfoFactory,
			CredFactory: fakeCredFactory,
		}
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

	Describe("RunWithOptions", func() {
		var output []string

		JustBeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				callCliCommandPlugin.RunWithOptions(fakeCliConnection, &options.Options{AppName: "app1"})
			})
		})

		Context("when there is an error getting the app model", func() {
			BeforeEach(func() {
				fakeAppFactory.GetReturns(app.App{}, errors.New("App not found"))
			})

			It("prints the error", func() {
				Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
				Expect(output).To(ContainSubstrings([]string{"App not found"}))
			})

			It("does not attempt to acquire endpoint info", func() {
				Expect(fakeInfoFactory.GetCallCount()).To(Equal(0))
			})
		})

		Context("when the app model is successfully acquired", func() {
			BeforeEach(func() {
				fakeAppFactory.GetReturns(app.App{}, nil)
				fakeInfoFactory.GetReturns(info.Info{}, nil)
			})

			It("gets the ssh endpoint information from /v2/info", func() {
				Expect(fakeInfoFactory.GetCallCount()).To(Equal(1))
			})

			Context("when getting the endpoint info fails", func() {
				BeforeEach(func() {
					fakeInfoFactory.GetReturns(info.Info{}, errors.New("woops"))
				})

				It("prints the error", func() {
					Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
					Expect(output).To(ContainSubstrings([]string{"woops"}))
				})
			})
		})

		Context("when getting the app model and endpoint info are successful", func() {
			BeforeEach(func() {
				fakeAppFactory.GetReturns(app.App{}, nil)
				fakeInfoFactory.GetReturns(info.Info{}, nil)
				fakeCredFactory.GetReturns(credential.Credential{}, nil)
			})

			It("gets the current oauth token credential", func() {
				Expect(fakeInfoFactory.GetCallCount()).To(Equal(1))
			})

			Context("when getting the credential fails", func() {
				BeforeEach(func() {
					fakeInfoFactory.GetReturns(info.Info{}, errors.New("woops"))
				})

				It("prints the error", func() {
					Expect(fakeAppFactory.GetCallCount()).To(Equal(1))
					Expect(output).To(ContainSubstrings([]string{"woops"}))
				})
			})
		})

		Context("when the app, endpoint, and credential are acquired", func() {
			BeforeEach(func() {
				app := app.App{
					Guid:      "app-guid",
					EnableSSH: true,
					Diego:     true,
					State:     "STARTED",
				}

				info := info.Info{
					SSHEndpoint:            "ssh.example.com:1234",
					SSHEndpointFingerprint: "fingerprint",
				}

				cred := credential.Credential{
					Token: "bearer token",
				}

				fakeAppFactory.GetReturns(app, nil)
				fakeInfoFactory.GetReturns(info, nil)
				fakeCredFactory.GetReturns(cred, nil)
			})

			It("dials the ssh endpoint", func() {
			})
		})
	})
})
