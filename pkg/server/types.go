package server

import (
	"go_redis/pkg/keyval"
	"go_redis/pkg/peer"
	"net"
)

type Config struct {
	ListenAddr string
}

type Message struct {
	Data []byte
	Peer *peer.Peer
}

type Server struct {
	Config
	peers     map[*peer.Peer]bool
	ln        net.Listener
	addPeerCh chan *peer.Peer
	quitCh    chan struct{}
	MsgCh     chan Message
	//
	kv *keyval.KV
}
