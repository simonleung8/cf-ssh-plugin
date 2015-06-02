package main_test

import (
	"errors"
	"net"

	"github.com/cloudfoundry-incubator/diego-ssh/authenticators/fake_authenticators"
	"github.com/cloudfoundry-incubator/diego-ssh/daemon"
	"github.com/cloudfoundry-incubator/diego-ssh/handlers"
	"github.com/cloudfoundry-incubator/diego-ssh/server"
	"github.com/cloudfoundry/cli/plugin/fakes"
	"github.com/pivotal-golang/lager"
	"github.com/pivotal-golang/lager/lagertest"
	"github.com/sykesm/cf-ssh-plugin"
	"github.com/sykesm/cf-ssh-plugin/models/app"
	"github.com/sykesm/cf-ssh-plugin/models/app/app_fakes"
	"github.com/sykesm/cf-ssh-plugin/models/credential"
	"github.com/sykesm/cf-ssh-plugin/models/credential/credential_fakes"
	"github.com/sykesm/cf-ssh-plugin/models/info"
	"github.com/sykesm/cf-ssh-plugin/models/info/info_fakes"
	"github.com/sykesm/cf-ssh-plugin/options"
	"golang.org/x/crypto/ssh"

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
		var (
			output []string
			opts   *options.Options
		)

		BeforeEach(func() {
			opts = &options.Options{}
		})

		JustBeforeEach(func() {
			output = io_helpers.CaptureOutput(func() {
				callCliCommandPlugin.RunWithOptions(fakeCliConnection, opts)
			})
		})

		Context("when there is an error getting the app model", func() {
			BeforeEach(func() {
				opts = &options.Options{
					AppName: "app1",
				}

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
			var (
				logger              lager.Logger
				daemonAuthenticator *fake_authenticators.FakePasswordAuthenticator

				sshInfo *info.Info

				sshDaemonListener net.Listener
				sshDaemon         *daemon.Daemon
				sshDaemonServer   *server.Server
			)

			BeforeEach(func() {
				logger = lagertest.NewTestLogger("test")

				var err error
				sshDaemonListener, err = net.Listen("tcp", "127.0.0.1:0")
				Expect(err).NotTo(HaveOccurred())

				daemonAuthenticator = &fake_authenticators.FakePasswordAuthenticator{}
				daemonAuthenticator.AuthenticateReturns(&ssh.Permissions{}, nil)

				daemonSSHConfig := &ssh.ServerConfig{}
				daemonSSHConfig.PasswordCallback = daemonAuthenticator.Authenticate
				daemonSSHConfig.AddHostKey(TestHostKey)

				sshDaemon = daemon.New(
					logger.Session("sshd"),
					daemonSSHConfig,
					map[string]handlers.GlobalRequestHandler{},
					map[string]handlers.NewChannelHandler{},
				)

				sshDaemonServer = server.NewServer(logger, "127.0.0.1:0", sshDaemon)
				sshDaemonServer.SetListener(sshDaemonListener)
				go sshDaemonServer.Serve()

				sshInfo = &info.Info{
					SSHEndpoint:            sshDaemonListener.Addr().String(),
					SSHEndpointFingerprint: TestHostKeyFingerprint,
				}

				app := app.App{
					Guid:      "app-guid",
					EnableSSH: true,
					Diego:     true,
					State:     "STARTED",
				}

				cred := credential.Credential{
					Token: "bearer token",
				}

				opts = &options.Options{
					AppName:  "app1",
					Instance: 2,
				}

				fakeAppFactory.GetReturns(app, nil)
				fakeCredFactory.GetReturns(cred, nil)
				fakeInfoFactory.GetStub = func() (info.Info, error) {
					return *sshInfo, nil
				}
			})

			AfterEach(func() {
				sshDaemonServer.Shutdown()
			})

			It("dials the ssh endpiont with the correct user and password", func() {
				Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(1))

				cmd, password := daemonAuthenticator.AuthenticateArgsForCall(0)
				Expect(cmd.User()).To(Equal("cf:app-guid/2"))
				Expect(password).To(BeEquivalentTo("bearer token"))
			})

			Context("when the SHA1 fingerprint does not match", func() {
				BeforeEach(func() {
					sshInfo.SSHEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
				})

				It("complains loudly with 'Host fingerprint does not match'", func() {
					Expect(output).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Host fingerprint does not match"},
					))
				})

				It("does not attempt to authenticate", func() {
					Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(0))
				})
			})

			Context("when the MD5 fingerprint does not match", func() {
				BeforeEach(func() {
					sshInfo.SSHEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
				})

				It("complains loudly with 'Host fingerprint does not match'", func() {
					Expect(output).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Host fingerprint does not match"},
					))
				})

				It("does not attempt to authenticate", func() {
					Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(0))
				})
			})

			Context("when the fingerprint length doesn't make sense", func() {
				BeforeEach(func() {
					sshInfo.SSHEndpointFingerprint = "garbage"
				})

				It("complains with message 'invalid fingerprint format'", func() {
					Expect(output).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"invalid fingerprint format"},
					))
				})

				It("does not attempt to authenticate", func() {
					Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(0))
				})
			})

			Context("when --skip-host-validation is provided", func() {
				BeforeEach(func() {
					sshInfo.SSHEndpointFingerprint = "00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00:00"
					opts.SkipHostValidation = true
				})

				It("ignores the host key", func() {
					Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(1))

					cmd, password := daemonAuthenticator.AuthenticateArgsForCall(0)
					Expect(cmd.User()).To(Equal("cf:app-guid/2"))
					Expect(password).To(BeEquivalentTo("bearer token"))
				})
			})

			Context("when no fingerprint is present at /v2/info", func() {
				BeforeEach(func() {
					sshInfo.SSHEndpointFingerprint = ""
				})

				It("successfully authenticates", func() {
					Expect(daemonAuthenticator.AuthenticateCallCount()).To(Equal(1))

					cmd, password := daemonAuthenticator.AuthenticateArgsForCall(0)
					Expect(cmd.User()).To(Equal("cf:app-guid/2"))
					Expect(password).To(BeEquivalentTo("bearer token"))
				})
			})

		})
	})
})
