package service

import (
	"log"
	"net"

	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const MaxSegmentSize = (1 << 16) - 1 // largest possible UDP datagram

type TunnelService struct {
	remoteConn            *net.UDPConn
	localChan, remoteChan chan *proto.Packet
	password              string
	proto.UnimplementedTunnelServiceServer
}

func NewTunnel(remoteConn *net.UDPConn, password string) *TunnelService {
	svc := &TunnelService{
		remoteChan: make(chan *proto.Packet),
		localChan:  make(chan *proto.Packet),
		remoteConn: remoteConn,
		password:   password,
	}
	go svc.handleRemoteConn()
	return svc
}

func (t *TunnelService) handleRemoteConn() error {

	go func() {

		for {
			p := <-t.localChan
			if p.Body != nil {
				if _, err := t.remoteConn.Write(p.Body); err != nil {
					log.Println(err)
				}
			}
		}

	}()

	for {

		buf := make([]byte, MaxSegmentSize)
		n, err := t.remoteConn.Read(buf[:])
		if err != nil {
			log.Fatal(err)
		}

		t.remoteChan <- &proto.Packet{
			Body: buf[:n],
		}

	}

}

func (t *TunnelService) Close() error {
	return t.remoteConn.Close()
}

func (t *TunnelService) Connect(stream proto.TunnelService_ConnectServer) error {

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

	log.Println("new connection: server_ready")

	go func() {

		for {
			p := <-t.remoteChan
			if p.Body != nil {
				if err := stream.Send(p); err != nil {
					log.Println(err)
				}
			}
		}

	}()

	for {

		req, err := stream.Recv()
		if err != nil {
			return errors.Wrapf(err, "can't receive message")
		}

		//log.Println(fmt.Sprintf("new packet: len[%d]", len(req.Body)))
		t.localChan <- req

	}

}
