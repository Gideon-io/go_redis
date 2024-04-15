package server

import (
	"fmt"
	"go_redis/pkg/keyval"
	"go_redis/pkg/proto"
	"log/slog"
	"net"
)

// create a new server with the given configuration
func NewServer(cfg Config) *Server {
	//if no listen address is provided, use the default
	if len(cfg.ListenAddr) == 0 {
		//cfg.ListenAddr = defaultListenAddr
		cfg.ListenAddr = ":5001"
	}
	//return a new server instance
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		MsgCh:     make(chan Message),
		kv:        keyval.NewKeyVal(),
	}
}

// start the server
func (s *Server) Start() error {
	//create a new listener
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	//store the listener
	s.ln = ln

	go s.loop()

	slog.Info("server started", "listenAddr", s.ListenAddr)

	return s.acceptLoop()

}

// handleMessage parses the message and executes the command
func (s *Server) handleMessage(msg Message) error {
	cmd, err := proto.ParseCommand(string(msg.Data))
	if err != nil {
		return err

	}
	switch v := cmd.(type) {
	case proto.SetCommand:
		return s.kv.Set(v.Key, v.Val)
	case proto.GetCommand:
		val, ok := s.kv.Get(v.Key)
		if !ok {
			return fmt.Errorf("key not found: %s", v.Key)
		}
		_, err := msg.Peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
		}
	}
	return nil
}

// loop is the main server loop that handles incoming messages and peer connections
func (s *Server) loop() {
	for {
		select {
		case msg := <-s.MsgCh:
			if err := s.handleMessage(msg); err != nil {
				slog.Error("raw message error", "err", err)
			}
		case <-s.quitCh:
			return
		case peer := <-s.addPeerCh:
			s.peers[peer] = true
		}
	}
}

// acceptLoop listens for incoming connections and starts a goroutine to handle each one
func (s *Server) acceptLoop() error {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accpet error", "err", err)
			continue
		}
		go s.handleConn(conn)
	}
}

// handleConn creates a new peer and adds it to the server
func (s *Server) handleConn(conn net.Conn) {
	peer := NewPeer(conn, s.MsgCh)
	s.addPeerCh <- peer
	if err := peer.ReadLoop(); err != nil {
		slog.Error("read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}
