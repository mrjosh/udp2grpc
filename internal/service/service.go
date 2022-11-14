package service

import (
	"fmt"
	"strings"

	"github.com/mrjosh/udp2grpc/internal/config"
	"github.com/mrjosh/udp2grpc/proto"
	"github.com/pkg/errors"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type TunnelService struct {
	logger    *logrus.Logger
	conf      *config.ConfMap
	peerspoll *Peers
	proto.UnimplementedTunnelServiceServer
}

func NewTunnel(logger *logrus.Logger, conf *config.ConfMap) *TunnelService {
	return &TunnelService{
		logger:    logger,
		conf:      conf,
		peerspoll: NewPeers(),
	}
}

func (t *TunnelService) Connect(stream proto.TunnelService_ConnectServer) error {

	//Create new Peer
	peer, err := t.authenticate(stream)
	if err != nil {
		return err
	}

	// add current peer to PeersPoll
	t.peerspoll.Add(peer)
	defer t.peerspoll.Remove(peer)

	return peer.Handle()
}

func (t *TunnelService) authenticate(stream proto.TunnelService_ConnectServer) (*Peer, error) {

	headers, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return nil, errors.New("could not get metadatas from context")
	}

	privatekey := headers.Get("privatekey")
	if len(privatekey) == 0 {
		return nil, errors.New("privatekey is required")
	}

	peerConf, err := t.conf.Server.FindPeer(privatekey[0])
	if err != nil {
		return nil, err
	}

	if err := t.IsAvailableFromThisSource(stream, peerConf); err != nil {
		return nil, err
	}

	peer, err := NewPeer(stream, t.logger, peerConf)
	if err != nil {
		return nil, err
	}

	return peer, nil
}

func (t *TunnelService) IsAvailableFromThisSource(stream proto.TunnelService_ConnectServer, peerConf *config.PeerConfMap) error {

	grpcPeer, ok := peer.FromContext(stream.Context())
	if !ok {
		return fmt.Errorf("could not get peer info")
	}

	addrStr := grpcPeer.Addr.String()
	addrArr := strings.Split(addrStr, ":")

	if len(addrArr) < 1 {
		return fmt.Errorf("could not decode peer.IPAddress")
	}
	addr := addrArr[0]

	for _, s := range peerConf.AvailableFrom {
		rng := ipaddr.NewIPAddressString(s).GetAddress()
		addr := ipaddr.NewIPAddressString(fmt.Sprintf("%s/32", addr)).GetAddress()
		if rng == nil && addr == nil {
			break
		}
		if rng.Contains(addr) {
			return nil
		}
	}

	return fmt.Errorf("permission denied from this source ip")
}

func (t *TunnelService) Close() error {
	return t.peerspoll.Close()
}
