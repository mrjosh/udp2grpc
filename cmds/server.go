package cmds

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/mrjosh/udp2grpc/internal/config"
	"github.com/mrjosh/udp2grpc/internal/service"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

type NewServerFlags struct {
	configfile string
}

func newServerCommand() *cobra.Command {
	cFlags := new(NewServerFlags)
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a udp2grpc tcp/server",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.configfile == "" {
				return errors.New("server config-file is required")
			}

			conf, err := config.LoadFile(cFlags.configfile)
			if err != nil {
				return err
			}

			localaddr := strings.Split(conf.Server.Listen, ":")
			if len(localaddr) < 2 {
				return fmt.Errorf("Local server address should match ip:port pattern")
			}

			listener, err := net.Listen("tcp4", conf.Server.Listen)
			if err != nil {
				return fmt.Errorf("could not create tcp listener: %v", err)
			}

			opts := []grpc.ServerOption{}

			if conf.Server.TLS != nil {
				if !conf.Server.TLS.Insecure {

					if conf.Server.TLS.CertFile == "" {
						return errors.New("tls.cert_file is required in tls mode. set tls off with tls.insecure in your config file")
					}
					if conf.Server.TLS.KeyFile == "" {
						return errors.New("tls.key_file is required in tls mode. set tls off with tls.insecure in your config file")
					}

					tlsCredentials, err := loadTLSCredentials(conf.Server.TLS.CertFile, conf.Server.TLS.KeyFile)
					if err != nil {
						return err
					}

					opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(tlsCredentials)))
				}
			}

			server := grpc.NewServer(opts...)

			// Register binance services
			svc := service.NewTunnel(logger, conf)
			defer svc.Close()

			proto.RegisterTunnelServiceServer(server, svc)

			reflection.Register(server)

			logger.Info(fmt.Sprintf("server running on tcp:%s", conf.Server.Listen))
			if err := server.Serve(listener); err != nil {
				return fmt.Errorf("could not serve grpc.tcp.listener: %v", err)
			}

			return nil
		},
	}
	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.configfile, "config-file", "c", "", "Server config file")
	return cmd
}

func loadTLSCredentials(certFile, keyFile string) (*tls.Certificate, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &serverCert, nil
}
