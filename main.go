package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"time"

	"go_redis/client"
)

const defaultListenAddr = ":5001"

type Config struct {
	ListenAddr string
}

type Message struct {
	data []byte
	peer *Peer
}

type Server struct {
	Config
	peers     map[*Peer]bool
	ln        net.Listener
	addPeerCh chan *Peer
	quitCh    chan struct{}
	msgCh     chan Message
	//
	kv *KV
}

// create a new server with the given configuration
func NewServer(cfg Config) *Server {
	//if no listen address is provided, use the default
	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}
	//return a new server instance
	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan Message),
		kv:        NewKeyVal(),
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

func (s *Server) handleMessage(msg Message) error {
	cmd, err := parseCommand(string(msg.data))
	if err != nil {
		return err

	}
	switch v := cmd.(type) {
	case SetCommand:
		return s.kv.Set(v.key, v.val)
	case GetCommand:
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key not found: %s", v.key)
		}
		_, err := msg.peer.Send(val)
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
		case msg := <-s.msgCh:
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
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer
	if err := peer.readLoop(); err != nil {
		slog.Error("read error", "err", err, "remoteAddr", conn.RemoteAddr())
	}
}

func main() {
	server := NewServer(Config{})
	go func() {

		log.Fatal(server.Start())
	}()
	//sleep so the server has time to start
	time.Sleep(time.Second)

	c := client.New("localhost:5001")
	for i := 0; i < 10; i++ {

		if err := c.Set(context.Background(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i)); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second)
		val, err := c.Get(context.Background(), fmt.Sprintf("foo_%d", i))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Got this back =>", val)

	}

}
