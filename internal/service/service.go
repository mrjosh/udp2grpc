package service

import (
	"fmt"
	"log"
	"net"

	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
)

const MaxSegmentSize = (1 << 16) - 1 // largest possible UDP datagram

type VPNService struct {
	remoteConn            *net.UDPConn
	localChan, remoteChan chan *proto.Packet
	proto.UnimplementedVPNServiceServer
}

func NewVPNService(remoteConn *net.UDPConn) *VPNService {
	svc := &VPNService{
		remoteChan: make(chan *proto.Packet),
		localChan:  make(chan *proto.Packet),
		remoteConn: remoteConn,
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

		log.Println(fmt.Sprintf("new packet: len[%d]", len(req.Body)))
		v.localChan <- req

	}

}
