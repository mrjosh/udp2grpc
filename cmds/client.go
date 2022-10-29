package cmds

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mrjosh/udp2grpc/internal/client"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type NewClientFlags struct {
	insecure                     bool
	localaddr, remoteaddr        string
	certFile, serverNameOverride string
	password                     string
}

func newClientCommand() *cobra.Command {

	log.SetFlags(log.Lshortfile)
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

			opts := []grpc.DialOption{}
			if cFlags.insecure {
				opts = append(opts, grpc.WithInsecure())
			}

			if !cFlags.insecure {

				if cFlags.certFile == "" {
					return errors.New("--tls-cert-file flag is required in tls mode. turn off tls mode with --insecure flag")
				}

				creds, err := credentials.NewClientTLSFromFile(cFlags.certFile, cFlags.serverNameOverride)
				if err != nil {
					log.Fatalln(err)
				}
				opts = append(opts, grpc.WithTransportCredentials(creds))
			}

			conn, err := grpc.Dial(cFlags.remoteaddr, opts...)
			if err != nil {
				return fmt.Errorf("did not connect: %v", err)
			}

			c := proto.NewTunnelServiceClient(conn)

			log.Println(fmt.Sprintf("connecting to tcp:%s", cFlags.remoteaddr))

			callOpts := grpc.EmptyCallOption{}

			md := metadata.New(map[string]string{
				"password": cFlags.password,
			})

			ctx := metadata.NewOutgoingContext(context.Background(), md)
			stream, err := c.Connect(ctx, callOpts)
			if err != nil {
				return err
			}

			log.Println(fmt.Sprintf("connected to tcp:%s client_ready", cFlags.remoteaddr))

			ic, err := client.NewClient(cFlags.localaddr, stream)
			if err != nil {
				return err
			}
			defer ic.Close()

			return ic.Listen()
		},
	}

	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.remoteaddr, "remote-address", "r", "", "Server remote address")
	cmd.Flags().StringVarP(&cFlags.localaddr, "local-address", "l", "", "Local server address")
	cmd.Flags().StringVarP(&cFlags.certFile, "tls-cert-file", "c", "", "Server TLS certificate file")
	cmd.Flags().StringVarP(&cFlags.serverNameOverride, "tls-server-name", "o", "", "TLS server name override")
	cmd.Flags().BoolVarP(&cFlags.insecure, "insecure", "I", false, "Connect to server without tls")
	cmd.Flags().StringVarP(&cFlags.password, "password", "p", "", "Server password")
	return cmd
}
