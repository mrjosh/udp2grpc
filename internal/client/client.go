package client

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/mrjosh/udp2grpc/internal/service"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
)

type Client struct {
	localConn             *net.UDPConn
	remoteStream          proto.VPNService_ConnectClient
	localChan, remoteChan chan *proto.Packet
	localConnAddr         net.Addr
}

func (c *Client) Close() error {
	return c.remoteStream.CloseSend()
}

func NewClient(localAddress string, remoteStream proto.VPNService_ConnectClient) (*Client, error) {
	log.Println(fmt.Sprintf("create a new local connection on udp:%s", localAddress))
	localConn, err := createNewLocalUDPListener(localAddress)
	if err != nil {
		return nil, err
	}
	c := &Client{
		remoteStream: remoteStream,
		localConn:    localConn,
		localChan:    make(chan *proto.Packet),
		remoteChan:   make(chan *proto.Packet),
	}
	go c.handleLocalConn()
	return c, nil
}

func (c *Client) Listen() error {

	go func() {

		for {
			p := <-c.localChan
			if p.Body != nil {
				if err := c.remoteStream.Send(p); err != nil {
					log.Println(err)
				}
			}
		}

	}()

	for {

		req, err := c.remoteStream.Recv()
		if err != nil {
			return errors.Wrapf(err, "can't receive message")
		}

		//log.Println(fmt.Sprintf("new packet: len[%d]", len(req.Body)))
		c.remoteChan <- req

	}

}

func createNewLocalUDPListener(address string) (*net.UDPConn, error) {

	local := strings.Split(address, ":")
	if len(local) < 2 {
		log.Fatal(errors.New("listen flag should contains ip:port"))
	}

	rport, err := strconv.Atoi(local[1])
	if err != nil {
		log.Fatal(errors.New("listen flag should contains ip:port"))
	}

	localConn, err := net.ListenUDP(
		"udp4",
		&net.UDPAddr{
			IP:   net.ParseIP(local[0]),
			Port: rport,
		},
	)

	if err != nil {
		return nil, err
	}

	return localConn, nil
}

func (c *Client) handleLocalConn() error {

	var laddr *net.Addr

	go func() {

		for {

			p := <-c.remoteChan
			if p.Body != nil && laddr != nil {
				if _, err := c.localConn.WriteTo(p.Body[:], *laddr); err != nil {
					log.Println(err)
				}
			}

		}

	}()

	for {

		buf := make([]byte, service.MaxSegmentSize)
		n, localAddr, err := c.localConn.ReadFrom(buf[:])
		if err != nil {
			log.Fatal(err)
		}
		laddr = &localAddr

		c.localChan <- &proto.Packet{
			Body: buf[:n],
		}

	}
}
