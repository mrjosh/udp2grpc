package cmds

import (
	"context"
	"fmt"
	"log"

	"github.com/mrjosh/udp2grpc/internal/client"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type NewClientFlags struct {
	localaddr, remoteaddr        string
	localport, remoteport        int
	insecure                     bool
	certFile, serverNameOverride string
}

func newClientCommand() *cobra.Command {

	log.SetFlags(log.Lshortfile)
	cFlags := new(NewClientFlags)

	cmd := &cobra.Command{
		Use:   "client",
		Short: "Start a udp2grpc tcp/client",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.remoteaddr == "" {
				return errors.New("server remote address is required. try 'utg client --address domain.tld'")
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

			addr := fmt.Sprintf("%s:%d", cFlags.remoteaddr, cFlags.remoteport)

			conn, err := grpc.Dial(addr, opts...)
			if err != nil {
				return fmt.Errorf("did not connect: %v", err)
			}

			c := proto.NewVPNServiceClient(conn)

			log.Println(fmt.Sprintf("Connecting to tcp:%s", addr))

			callOpts := grpc.EmptyCallOption{}
			stream, err := c.Connect(context.Background(), callOpts)
			if err != nil {
				return err
			}

			log.Println(fmt.Sprintf("Connected to tcp:%s", addr))

			ic, err := client.NewClient(fmt.Sprintf("%s:%d", cFlags.localaddr, cFlags.localport), stream)
			if err != nil {
				return err
			}
			defer ic.Close()

			return ic.Listen()
		},
	}

	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.remoteaddr, "remote-address", "r", "", "Server remote address")
	cmd.Flags().IntVarP(&cFlags.remoteport, "port", "p", 52935, "Server tcp port")
	cmd.Flags().StringVarP(&cFlags.localaddr, "local-address", "l", "", "Local server address")
	cmd.Flags().IntVarP(&cFlags.localport, "local-port", "P", 52935, "Local server port")
	cmd.Flags().StringVarP(&cFlags.certFile, "tls-cert-file", "c", "", "Server TLS certificate file")
	cmd.Flags().StringVarP(&cFlags.serverNameOverride, "tls-server-name", "o", "", "TLS server name override")
	cmd.Flags().BoolVarP(&cFlags.insecure, "insecure", "k", false, "Connect to server without tls")
	return cmd
}
