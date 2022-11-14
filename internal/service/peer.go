package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/mrjosh/udp2grpc/internal/config"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
)

const MaxSegmentSize = (1 << 16) - 1 // largest possible TCP datagram

type Peer struct {
	id           string
	logger       *logrus.Logger
	stream       proto.TunnelService_ConnectServer
	localAddr    net.Addr
	localChan    chan *proto.Packet
	remoteChan   chan *proto.Packet
	remoteConn   *net.UDPConn
	peerLogEntry *logrus.Entry
}

func NewPeer(stream proto.TunnelService_ConnectServer, logger *logrus.Logger, peerConf *config.PeerConfMap) (*Peer, error) {
	// create a new uuid for peer connection
	id := uuid.New().String()
	// get the current peer IPAddress
	grpcPeer, ok := peer.FromContext(stream.Context())
	if !ok {
		return nil, fmt.Errorf("could not get peer info")
	}
	logEntry := logger.WithFields(logrus.Fields{
		"id":   id,
		"addr": grpcPeer.Addr.String(),
		"peer": peerConf.Name,
	})
	p := &Peer{
		id:           id,
		logger:       logger,
		stream:       stream,
		localAddr:    grpcPeer.Addr,
		localChan:    make(chan *proto.Packet),
		remoteChan:   make(chan *proto.Packet),
		peerLogEntry: logEntry,
	}
	logEntry.Info("new connection established")
	if err := p.createRemoteConnection(stream.Context(), peerConf.Remote); err != nil {
		return nil, err
	}
	return p, nil
}

func (peer *Peer) createRemoteConnection(ctx context.Context, address string) error {

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

	peer.remoteConn = remoteConn
	go peer.handleRemoteConn(ctx)

	return nil
}

func (peer *Peer) handleRemoteConn(ctx context.Context) error {

	go func() {

		for {

			select {
			case p, ok := <-peer.localChan:
				if p != nil && ok {

					switch p.Type {
					case proto.PACKET_TYPE_PING:
						//peer.peerLogEntry.Infof("[%s] receiving Ping Packet")
						peer.remoteChan <- &proto.Packet{
							Type: proto.PACKET_TYPE_PONG,
						}
					case proto.PACKET_TYPE_BODY:
						go peer.remoteConn.Write(p.Body[:])
					}

				}
			case <-ctx.Done():
				peer.remoteConn.Close()
				return
			}

		}

	}()

	for {

		buf := make([]byte, MaxSegmentSize)
		n, err := peer.remoteConn.Read(buf[:])
		if err != nil {
			return err
		}

		peer.remoteChan <- &proto.Packet{
			Body: buf[:n],
			Type: proto.PACKET_TYPE_BODY,
		}

	}

}

func (peer *Peer) Handle() error {

	go func() {

		for {
			select {
			case p, ok := <-peer.remoteChan:
				if p != nil && ok {
					if err := peer.stream.Send(p); err != nil {
						peer.peerLogEntry.Println(err)
					}
				}
			case <-peer.stream.Context().Done():
				peer.remoteConn.Close()
				return
			}
		}

	}()

	for {

		req, err := peer.stream.Recv()
		if err != nil {
			peer.peerLogEntry.Errorf("connection error. %v", err)
			return errors.Wrapf(err, "can't receive message")
		}
		peer.localChan <- req

	}

}

func (peer *Peer) Close() {
	peer.remoteConn.Close()
	peer.stream.Context().Done()
	peer.peerLogEntry.Warnf("connection closed.")
}

type Peers struct {
	mu    sync.Mutex
	peers map[string]*Peer
}

func NewPeers() *Peers {
	return &Peers{
		peers: make(map[string]*Peer, 0),
	}
}

func (p *Peers) Add(peer *Peer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers[peer.localAddr.String()] = peer
}

func (p *Peers) Remove(peer *Peer) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers[peer.localAddr.String()].Close()
	delete(p.peers, peer.localAddr.String())
}

func (p *Peers) Close() error {
	return nil
}
