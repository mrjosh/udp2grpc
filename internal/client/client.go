package client

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/mrjosh/udp2grpc/internal/service"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
)

type Client struct {
	ctx                   context.Context
	logger                *logrus.Logger
	localConn             *net.UDPConn
	remoteStream          proto.TunnelService_ConnectClient
	remoteConn            *grpc.ClientConn
	localChan, remoteChan chan *proto.Packet
	localConnAddr         net.Addr
	reconnect             chan bool
	done                  chan bool
	password              string
	remoteaddr            string
	persistentKeepalive   int64
}

func (c *Client) Close() error {
	return c.remoteStream.CloseSend()
}

func (c *Client) ProcessListen() error {

	go c.listen()

	for {
		select {
		case <-c.reconnect:
			// TODO: reconnect to upstream
			return fmt.Errorf("TODO: try to reconnect on failure")
		case <-c.done:
			return nil
		}
	}

}

func (c *Client) waitUntilReady() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return c.remoteConn.WaitForStateChange(ctx, connectivity.Ready)
}

func NewClient(ctx context.Context, logger *logrus.Logger, remoteConn *grpc.ClientConn, localaddr, remoteaddr, password string, pka int64) (*Client, error) {
	logger.Println(fmt.Sprintf("create a new local connection on udp:%s", localaddr))
	localConn, err := createNewLocalUDPListener(localaddr)
	if err != nil {
		return nil, err
	}
	c := &Client{
		ctx:                 ctx,
		logger:              logger,
		remoteaddr:          remoteaddr,
		password:            password,
		remoteConn:          remoteConn,
		localConn:           localConn,
		localChan:           make(chan *proto.Packet),
		remoteChan:          make(chan *proto.Packet),
		done:                make(chan bool),
		reconnect:           make(chan bool),
		persistentKeepalive: pka,
	}
	go c.handleLocalConn()
	return c, nil
}

func (c *Client) SetLogger(logger *logrus.Logger) {
	c.logger = logger
}

func (c *Client) newStream(ctx context.Context, conn *grpc.ClientConn) (proto.TunnelService_ConnectClient, error) {

	tunnel := proto.NewTunnelServiceClient(conn)
	md := metadata.New(map[string]string{
		"password": c.password,
	})

	callOpts := []grpc.CallOption{}
	c.logger.Infof("connecting to tcp:%s", c.remoteaddr)

	mdCtx := metadata.NewOutgoingContext(ctx, md)
	stream, err := tunnel.Connect(mdCtx, callOpts...)
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (c *Client) pingPongKeepAlive() {

	ticker := time.NewTicker(time.Second * time.Duration(c.persistentKeepalive))
	for {
		<-ticker.C
		c.remoteStream.Send(&proto.Packet{
			Type: proto.PACKET_TYPE_PING,
		})
	}

}

func (c *Client) listen() error {

	stream, err := c.newStream(c.ctx, c.remoteConn)
	if err != nil {
		return err
	}

	c.remoteStream = stream
	c.logger.Infof("connected to tcp:%s client_ready", c.remoteaddr)

	go c.pingPongKeepAlive()

	go func() {

		for {

			select {
			case p, ok := <-c.localChan:
				if p != nil && ok {
					if err := c.remoteStream.Send(p); err != nil {
						c.logger.Error(err)
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
			c.reconnect <- true
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
