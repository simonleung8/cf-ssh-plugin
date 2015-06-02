package main_test

import (
	"github.com/cloudfoundry-incubator/diego-ssh/helpers"
	"github.com/cloudfoundry-incubator/diego-ssh/keys"
	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"
	"golang.org/x/crypto/ssh"

	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	TestHostKey            ssh.Signer
	TestHostKeyFingerprint string
	TestPrivateKey         ssh.Signer
)

func TestDiegoSsh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DiegoSsh Suite")
}

// func TestSSHDaemon(t *testing.T) {
// 	RegisterFailHandler(Fail)
// 	RunSpecs(t, "Sshd Suite")
// }

var _ = BeforeSuite(func() {
	plugin_builder.BuildTestBinary(".", "ssh")

	hostKey, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
	Expect(err).NotTo(HaveOccurred())

	privateKey, err := keys.RSAKeyPairFactory.NewKeyPair(1024)
	Expect(err).NotTo(HaveOccurred())

	TestHostKey = hostKey.PrivateKey()
	TestHostKeyFingerprint = helpers.SHA1Fingerprint(TestHostKey.PublicKey())

	TestPrivateKey = privateKey.PrivateKey()

	// TestPrivatePem = privateKey.PEMEncodedPrivateKey()
	// TestPublicAuthorizedKey = privateKey.AuthorizedKey()
})
