package service

import (
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/sirupsen/logrus"
)

type TunnelService struct {
	logger     *logrus.Logger
	remoteaddr string
	password   string
	peers      *Peers
	proto.UnimplementedTunnelServiceServer
}

func NewTunnel(logger *logrus.Logger, remoteaddr, password string) *TunnelService {
	return &TunnelService{
		logger:     logger,
		remoteaddr: remoteaddr,
		password:   password,
		peers:      NewPeers(),
	}
}

func (t *TunnelService) Connect(stream proto.TunnelService_ConnectServer) error {

	// Create new Peer
	p, err := NewPeer(stream, t.logger, t.remoteaddr, t.password)
	if err != nil {
		return err
	}

	// add current peer to PeersPoll
	t.peers.Add(p)
	defer t.peers.Remove(p)

	return p.Handle()
}

func (t *TunnelService) Close() error {
	return t.peers.Close()
}
