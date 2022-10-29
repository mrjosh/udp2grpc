package service

import (
	"log"
	"net"

	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const MaxSegmentSize = (1 << 16) - 1 // largest possible UDP datagram

type VPNService struct {
	remoteConn            *net.UDPConn
	localChan, remoteChan chan *proto.Packet
	password              string
	proto.UnimplementedVPNServiceServer
}

func NewVPNService(remoteConn *net.UDPConn, password string) *VPNService {
	svc := &VPNService{
		remoteChan: make(chan *proto.Packet),
		localChan:  make(chan *proto.Packet),
		remoteConn: remoteConn,
		password:   password,
	}
	go svc.handleRemoteConn()
	return svc
}

func (v *VPNService) handleRemoteConn() error {

	go func() {

		for {
			p := <-v.localChan
			if p.Body != nil {
				if _, err := v.remoteConn.Write(p.Body); err != nil {
					log.Println(err)
				}
			}
		}

	}()

	for {

		buf := make([]byte, MaxSegmentSize)
		n, err := v.remoteConn.Read(buf[:])
		if err != nil {
			log.Fatal(err)
		}

		v.remoteChan <- &proto.Packet{
			Body: buf[:n],
		}

	}

}

func (v *VPNService) Close() error {
	return v.remoteConn.Close()
}

func (v *VPNService) Connect(stream proto.VPNService_ConnectServer) error {

	headers, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return errors.New("could not get metadatas from context")
	}

	password := headers.Get("password")
	if len(password) == 0 {
		return errors.New("password is required")
	}

	if password[0] != v.password {
		return errors.New("password incorrect")
	}

	log.Println("new connection: server_ready")

	go func() {

		for {
			p := <-v.remoteChan
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
		v.localChan <- req

	}

}
