package cmds

import (
	"fmt"
	"log"
	"os"

	"github.com/mrjosh/udp2grpc/internal/certificate"
	"github.com/spf13/cobra"
)

type NewGenCertificatesFlags struct {
	directory string
	domain    string
	serverip  string
}

func newGenCertificatesCommand() *cobra.Command {

	log.SetFlags(log.Lshortfile)
	cFlags := new(NewGenCertificatesFlags)

	cmd := &cobra.Command{
		Use:   "gen-certificates",
		Short: "Generate certificates for server and client",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.directory == "" {
				cFlags.directory = os.Getenv("PWD")
			}

			if cFlags.domain == "" && cFlags.serverip == "" {
				return fmt.Errorf("use --ip or --domain")
			}

			cert := certificate.Certificate{
				Subject: "CN=Josh",
				KeyType: certificate.KeyTypeRSA,
				KeySize: 1024,
			}

			if cFlags.domain != "" {
				cert.SubjectAltNames = append(cert.SubjectAltNames, fmt.Sprintf("DNS:%s", cFlags.domain))
			}

			if cFlags.serverip != "" {
				cert.SubjectAltNames = append(cert.SubjectAltNames, fmt.Sprintf("IP:%s", cFlags.serverip))
			}

			// generating certificate
			if err := cert.Generate(); err != nil {
				return err
			}

			// save certificate into crt,key files
			return cert.WritePEM(
				fmt.Sprintf("%s/server.crt", cFlags.directory),
				fmt.Sprintf("%s/server.key", cFlags.directory),
			)

		},
	}
	cmd.Flags().StringVarP(&cFlags.directory, "dir", "d", "", "Certificates directory")
	cmd.Flags().StringVarP(&cFlags.domain, "domain", "D", "", "Top-Level domain")
	cmd.Flags().StringVarP(&cFlags.serverip, "ip", "I", "", "Server IPAddress")
	return cmd
}
