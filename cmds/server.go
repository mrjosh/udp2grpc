package cmds

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mrjosh/udp2grpc/internal/service"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type NewServerFlags struct {
	localaddr, remoteaddr string
	insecure              bool
	certFile, keyFile     string
}

func newServerCommand() *cobra.Command {

	log.SetFlags(log.Lshortfile)
	cFlags := new(NewServerFlags)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a udp2grpc tcp/server",
		RunE: func(cmd *cobra.Command, args []string) error {

			if cFlags.remoteaddr == "" {
				return fmt.Errorf("Server remote address is required. try with flag 'utg server -r127.0.0.1:port '")
			}

			localaddr := strings.Split(cFlags.localaddr, ":")
			if len(localaddr) < 2 {
				return fmt.Errorf("Local server address should contain ip:port")
			}

			listener, err := net.Listen("tcp4", cFlags.localaddr)
			if err != nil {
				return fmt.Errorf("could not create tcp listener: %v", err)
			}

			kaParams := keepalive.ServerParameters{
				Time:    10 * time.Second,
				Timeout: 5 * time.Second,
			}

			enforcement := keepalive.EnforcementPolicy{
				MinTime:             3 * time.Second,
				PermitWithoutStream: true,
			}

			opts := []grpc.ServerOption{
				grpc.KeepaliveParams(kaParams),
				grpc.KeepaliveEnforcementPolicy(enforcement),
			}

			if !cFlags.insecure {

				if cFlags.certFile == "" {
					return errors.New("--tls-cert-file flag is required in tls mode. turn off tls mode with --insecure flag")
				}
				if cFlags.keyFile == "" {
					return errors.New("--tls-key-file flag is required in tls mode. turn off tls mode with --insecure flag")
				}

				tlsCredentials, err := loadTLSCredentials(cFlags.certFile, cFlags.keyFile)
				if err != nil {
					return err
				}

				opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(tlsCredentials)))
			}

			server := grpc.NewServer(opts...)

			remoteConn, err := createRemoteConnection(cFlags.remoteaddr)
			if err != nil {
				return err
			}

			// Register binance services
			svc := service.NewVPNService(remoteConn)
			defer svc.Close()

			proto.RegisterVPNServiceServer(server, svc)

			reflection.Register(server)

			log.Println(fmt.Sprintf("Server running in tcp:%s", cFlags.localaddr))
			if err := server.Serve(listener); err != nil {
				return fmt.Errorf("could not serve grpc.tcp.listener: %v", err)
			}

			return nil
		},
	}
	cmd.SuggestionsMinimumDistance = 1
	cmd.Flags().StringVarP(&cFlags.localaddr, "local-address", "l", "0.0.0.0:52935", "Local server address")
	cmd.Flags().StringVarP(&cFlags.remoteaddr, "remote-address", "r", "", "Remote address")
	cmd.Flags().StringVarP(&cFlags.certFile, "tls-cert-file", "c", "", "Server TLS certificate file")
	cmd.Flags().StringVarP(&cFlags.keyFile, "tls-key-file", "k", "", "Server TLS key file")
	cmd.Flags().BoolVarP(&cFlags.insecure, "insecure", "I", false, "Start the server without tls")
	return cmd
}

func createRemoteConnection(address string) (*net.UDPConn, error) {

	remote := strings.Split(address, ":")
	rport, err := strconv.Atoi(remote[1])
	if err != nil {
		log.Fatal(errors.New("listen flag should contains ip:port"))
	}

	remoteConn, err := net.DialUDP(
		"udp",
		&net.UDPAddr{},
		&net.UDPAddr{
			IP:   net.ParseIP(remote[0]),
			Port: rport,
		},
	)

	if err != nil {
		return nil, err
	}

	return remoteConn, nil
}

func loadTLSCredentials(certFile, keyFile string) (*tls.Certificate, error) {
	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	return &serverCert, nil
}
