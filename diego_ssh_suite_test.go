package main_test

import (
	"github.com/cloudfoundry-incubator/diego-ssh/server"
	"github.com/cloudfoundry/cli/testhelpers/plugin_builder"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDiegoSsh(t *testing.T) {
	RegisterFailHandler(Fail)

	plugin_builder.BuildTestBinary(".", "ssh")

	RunSpecs(t, "DiegoSsh Suite")
}

func NewServer(connectionHandler server.ConnectionHandler) *server.Server {
}
