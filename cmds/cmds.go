package cmds

import (
	"github.com/mrjosh/udp2grpc/internal/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	err         error
	logger      = logrus.New()
	versionInfo *version.BuildInfo
)

func Start(vi *version.BuildInfo, rootCmd *cobra.Command) error {

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetOutput(rootCmd.OutOrStderr())

	vi.BuildType = "Release"
	if vi.Branch == "develop" {
		vi.BuildType = "Nightly"
	}

	versionInfo = vi
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newServerCommand())
	rootCmd.AddCommand(newClientCommand())
	rootCmd.AddCommand(newGenCertificatesCommand())

	return rootCmd.Execute()
}
