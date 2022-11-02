package client

import (
	"context"
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
	ctx                   context.Context
	localConn             *net.UDPConn
	remoteStream          proto.TunnelService_ConnectClient
	localChan, remoteChan chan *proto.Packet
	localConnAddr         net.Addr
}

func (c *Client) Close() error {
	return c.remoteStream.CloseSend()
}

func NewClient(ctx context.Context, localAddress string, remoteStream proto.TunnelService_ConnectClient) (*Client, error) {
	log.Println(fmt.Sprintf("create a new local connection on udp:%s", localAddress))
	localConn, err := createNewLocalUDPListener(localAddress)
	if err != nil {
		return nil, err
	}
	c := &Client{
		ctx:          ctx,
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

			select {
			case p, ok := <-c.localChan:
				if p != nil && ok {
					if err := c.remoteStream.Send(p); err != nil {
						log.Println(err)
						return
					}
				}
			case <-c.ctx.Done():
				c.remoteStream.CloseSend()
				return
			}

		}

	}()

	for {

		req, err := c.remoteStream.Recv()
		if err != nil {
			return errors.Wrapf(err, "can't receive message")
		}
		c.remoteChan <- req

	}

}

func createNewLocalUDPListener(address string) (*net.UDPConn, error) {

	local := strings.Split(address, ":")
	if len(local) < 2 {
		return nil, errors.New("listen flag should contains ip:port")
	}

	rport, err := strconv.Atoi(local[1])
	if err != nil {
		return nil, errors.New("listen flag should contains ip:port")
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

			select {
			case p, ok := <-c.remoteChan:
				if laddr != nil && ok {

					switch p.Type {
					case proto.PACKET_TYPE_PING:
						c.localChan <- &proto.Packet{
							Type: proto.PACKET_TYPE_PONG,
						}
					case proto.PACKET_TYPE_BODY:
						go c.localConn.WriteTo(p.Body[:], *laddr)
					}

				}
			case <-c.ctx.Done():
				c.localConn.Close()
				return
			}

		}

	}()

	for {

		buf := make([]byte, service.MaxSegmentSize)
		n, localAddr, err := c.localConn.ReadFrom(buf[:])
		if err != nil {
			return err
		}

		laddr = &localAddr
		c.localChan <- &proto.Packet{
			Body: buf[:n],
			Type: proto.PACKET_TYPE_BODY,
		}

	}
}
