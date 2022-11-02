package service

import (
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const MaxSegmentSize = (1 << 16) - 1 // largest possible UDP datagram

type TunnelService struct {
	remoteConn            *net.UDPConn
	remoteAddr            string
	localChan, remoteChan chan *proto.Packet
	password              string
	proto.UnimplementedTunnelServiceServer
}

func NewTunnel(remoteAddr, password string) *TunnelService {
	return &TunnelService{
		remoteChan: make(chan *proto.Packet),
		localChan:  make(chan *proto.Packet),
		remoteAddr: remoteAddr,
		password:   password,
	}
}

func (t *TunnelService) createRemoteConnection(ctx context.Context, address string) error {

	remote := strings.Split(address, ":")
	rport, err := strconv.Atoi(remote[1])
	if err != nil {
		return errors.New("listen flag should contains [ip/domain]:port")
	}

	raddrs, err := net.LookupIP(remote[0])
	if err != nil {
		return err
	}

	if len(raddrs) == 0 {
		return fmt.Errorf("could not resolve domain [%s]", remote[0])
	}

	var raddr net.IP
	for _, addr := range raddrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			raddr = ipv4
			break
		}
	}

	if raddr == nil {
		raddr = raddrs[0]
	}

	remoteConn, err := net.DialUDP(
		"udp",
		&net.UDPAddr{},
		&net.UDPAddr{
			IP:   raddr,
			Port: rport,
		},
	)

	if err != nil {
		return err
	}

	t.remoteConn = remoteConn
	go t.handleRemoteConn(ctx)

	return nil
}

func (t *TunnelService) handleRemoteConn(ctx context.Context) error {

	go func() {

		for {

			select {
			case p, ok := <-t.localChan:
				if p != nil && ok {

					switch p.Type {
					case proto.PACKET_TYPE_PING:
						t.remoteChan <- &proto.Packet{
							Type: proto.PACKET_TYPE_PONG,
						}
					case proto.PACKET_TYPE_BODY:
						go t.remoteConn.Write(p.Body[:])
					}

				}
			case <-ctx.Done():
				t.remoteConn.Close()
				return
			}

		}

	}()

	for {

		buf := make([]byte, MaxSegmentSize)
		n, err := t.remoteConn.Read(buf[:])
		if err != nil {
			return err
		}

		t.remoteChan <- &proto.Packet{
			Body: buf[:n],
			Type: proto.PACKET_TYPE_BODY,
		}

	}

}

func (t *TunnelService) Close() error {
	return t.remoteConn.Close()
}

func (t *TunnelService) Connect(stream proto.TunnelService_ConnectServer) error {

	if err := t.authenticate(stream); err != nil {
		return err
	}

	if err := t.createRemoteConnection(stream.Context(), t.remoteAddr); err != nil {
		return err
	}

	log.Println("new connection: server_ready")

	go func() {

		for {
			select {
			case p, ok := <-t.remoteChan:
				if p != nil && ok {
					if err := stream.Send(p); err != nil {
						log.Println(err)
					}
				}
			case <-stream.Context().Done():
				t.remoteConn.Close()
				return
			}
		}

	}()

	for {

		req, err := stream.Recv()
		if err != nil {
			return errors.Wrapf(err, "can't receive message")
		}
		t.localChan <- req

	}

}

func (t *TunnelService) authenticate(stream proto.TunnelService_ConnectServer) error {

	headers, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("could not get metadatas from context")
	}

	password := headers.Get("password")
	if len(password) == 0 {
		return errors.New("password is required")
	}

	if password[0] != t.password {
		return errors.New("password incorrect")
	}

	return nil
}
