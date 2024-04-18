package server

import (
	"go_redis/pkg/keyval"
	"net"

	"github.com/fzzy/radix/redis/resp"
)

type Config struct {
	ListenAddr string
}

type Message struct {
	Data []byte
	Peer *Peer
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	MsgCh     chan Message
	//
	kv *keyval.KV
}

func hello() {
	resp.AppendArbitraryAsFlattenedStrings
}
