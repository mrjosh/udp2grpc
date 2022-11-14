package cmds

import (
	"fmt"
	"strings"
	"time"

	"github.com/mrjosh/udp2grpc/internal/client"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
)

type NewClientFlags struct {
	insecure                     bool
	localaddr, remoteaddr        string
	certFile, serverNameOverride string
	password                     string
	persistentKeepalive          int64
}

func newClientCommand() *cobra.Command {
	cFlags := new(NewClientFlags)
	cmd := &cobra.Command{
		Use:   "client",
		Short: "Start a udp2grpc tcp/client",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.password == "" {
				return errors.New("server password is required")
			}

			if cFlags.remoteaddr == "" {
				return errors.New("server remote address is required. try 'utg client -rdomain.tld:52935'")
			}

			remoteaddr := strings.Split(cFlags.remoteaddr, ":")
			if len(remoteaddr) < 2 {
				return fmt.Errorf("Remote server address should contain ip:port")
			}

			opts := []grpc.DialOption{
				grpc.WithBlock(),
			}

			if cFlags.insecure {
				opts = append(opts, grpc.WithInsecure())
			}

			if !cFlags.insecure {

				if cFlags.certFile == "" {
					return errors.New("--tls-cert-file flag is required in tls mode. turn off tls mode with --insecure flag")
				}

				creds, err := credentials.NewClientTLSFromFile(cFlags.certFile, cFlags.serverNameOverride)
				if err != nil {
					return err
				}
				opts = append(opts, grpc.WithTransportCredentials(creds))
			}

			remoteConn, err := grpc.DialContext(cmd.Context(), cFlags.remoteaddr, opts...)
			if err != nil {
				return fmt.Errorf("did not connect: %v", err)
			}

			ic, err := client.NewClient(
				cmd.Context(),
				logger,
				remoteConn,
				cFlags.localaddr,
				cFlags.remoteaddr,
				cFlags.password,
				cFlags.persistentKeepalive,
			)
			if err != nil {
				return err
			}
			defer ic.Close()

			return ic.ProcessListen()
		},
	}

	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.remoteaddr, "remote-address", "r", "", "Server remote address")
	cmd.Flags().StringVarP(&cFlags.localaddr, "local-address", "l", "", "Local server address")
	cmd.Flags().StringVarP(&cFlags.certFile, "tls-cert-file", "c", "", "Server TLS certificate file")
	cmd.Flags().StringVarP(&cFlags.serverNameOverride, "tls-server-name", "o", "", "TLS server name override")
	cmd.Flags().BoolVarP(&cFlags.insecure, "insecure", "I", false, "Connect to server without tls")
	cmd.Flags().StringVarP(&cFlags.password, "password", "p", "", "Server password")
	cmd.Flags().Int64VarP(&cFlags.persistentKeepalive, "PersistentKeepalive", "P", 30, "Persistent Keepalive")
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
