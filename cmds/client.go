package cmds

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrjosh/udp2grpc/internal/client"
	"github.com/mrjosh/udp2grpc/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type NewClientFlags struct {
	configfile string
}

func newClientCommand() *cobra.Command {
	cFlags := new(NewClientFlags)
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Start a udp2grpc tcp/client",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.configfile == "" {
				return errors.New("server config-file is required")
			}

			conf, err := config.LoadFile(cFlags.configfile)
			if err != nil {
				return err
			}

			if conf.Client.Remote == "" {
				return errors.New("server remote address is required")
			}

			remoteaddr := strings.Split(conf.Client.Remote, ":")
			if len(remoteaddr) < 2 {
				return fmt.Errorf("remote server address should contain ip:port")
			}

			opts := []grpc.DialOption{
				grpc.WithBlock(),
			}

			if conf.Client.TLS.Insecure {
				opts = append(opts, grpc.WithInsecure())
			}

			if conf.Client.TLS != nil {
				if !conf.Client.TLS.Insecure {

					if conf.Client.TLS.CertFile == "" {
						return errors.New("tls.cert_file is required in tls mode. set tls mode off with `tls.insecure` in your config file")
					}

					creds, err := credentials.NewClientTLSFromFile(conf.Client.TLS.CertFile, "")
					if err != nil {
						return err
					}
					opts = append(opts, grpc.WithTransportCredentials(creds))
				}
			}

			remoteConn, err := grpc.DialContext(cmd.Context(), conf.Client.Remote, opts...)
			if err != nil {
				return fmt.Errorf("did not connect: %v", err)
			}

			ic, err := client.NewClient(cmd.Context(), conf, logger, remoteConn)
			if err != nil {
				return err
			}
			defer ic.Close()

			return ic.Listen()
		},
	}

	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.configfile, "config-file", "c", "", "Client config file")
	return cmd
}

func reconnect(remoteaddr string, conn *grpc.ClientConn) error {

	logger.Warnf("reconnecting to tcp:%s", remoteaddr)

	if conn.GetState() != connectivity.Ready {
		ticker := time.NewTicker(time.Second * 5)
		for {
			<-ticker.C
			conn.Connect()
		}
	}

	return nil
}
