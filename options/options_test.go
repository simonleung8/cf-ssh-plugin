package options_test

import (
	"github.com/sykesm/cf-ssh-plugin/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Options", func() {
	var (
		opts       *options.Options
		args       []string
		parseError error
	)

	BeforeEach(func() {
		opts = &options.Options{}
		args = []string{}
		parseError = nil
	})

	JustBeforeEach(func() {
		parseError = opts.Parse(args)
	})

	Context("when no arguments are provided", func() {
		BeforeEach(func() {
			args = []string{}
		})

		It("returns a UsageError", func() {
			Expect(parseError).To(Equal(options.UsageError))
		})
	})

	Context("when no app name is provided", func() {
		BeforeEach(func() {
			args = []string{"-i", "3"}
		})

		It("returns a UsageError", func() {
			Expect(parseError).To(Equal(options.UsageError))
		})
	})

	Context("when an app name is provided", func() {
		Context("as the only argument", func() {
			BeforeEach(func() {
				args = []string{"App-1"}
			})

			It("populates the AppName field", func() {
				Expect(parseError).NotTo(HaveOccurred())
				Expect(opts.AppName).To(Equal("App-1"))
			})
		})

		Context("as the last argument", func() {
			BeforeEach(func() {
				args = []string{"-i", "3", "App-1"}
			})

			It("populates the AppName field", func() {
				Expect(parseError).NotTo(HaveOccurred())
				Expect(opts.AppName).To(Equal("App-1"))
			})
		})
	})

	Context("when --skip-host-validation is set", func() {
		BeforeEach(func() {
			args = []string{"app-name", "--skip-host-validation"}
		})

		It("disables host key validation", func() {
			Expect(parseError).ToNot(HaveOccurred())
			Expect(opts.SkipHostValidation).To(BeTrue())
		})

		Context("when --skip-host-validation=false is set", func() {
			BeforeEach(func() {
				args = []string{"app-name", "--skip-host-validation=false"}
			})

			It("disables host key validation", func() {
				Expect(parseError).ToNot(HaveOccurred())
				Expect(opts.SkipHostValidation).To(BeFalse())
			})
		})
	})

	Context("when an -i flag is provided", func() {
		BeforeEach(func() {
			args = []string{"app-name"}
		})

		Context("without an argument", func() {
			BeforeEach(func() {
				args = append(args, "-i")
			})

			It("returns an error", func() {
				Expect(parseError).To(MatchError("No value provided for flag: -i"))
			})
		})

		Context("with a positive integer argument", func() {
			BeforeEach(func() {
				args = append(args, "-i", "3")
			})

			It("populates the Instance field", func() {
				Expect(parseError).NotTo(HaveOccurred())
				Expect(opts.Instance).To(Equal(3))
			})
		})

		Context("with a negative integer argument", func() {
			BeforeEach(func() {
				args = append(args, "-i", "-3")
			})

			It("populates the Instance field", func() {
				Expect(parseError).To(MatchError("Value for flag 'i' must not be negative"))
			})
		})

		Context("with a non-numeric argument", func() {
			BeforeEach(func() {
				args = append(args, "-i", "three")
			})

			It("returns an error", func() {
				Expect(parseError).To(MatchError("Value for flag 'i' must be integer"))
			})
		})

	})
})
