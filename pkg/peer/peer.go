package peer

import (
	"go_redis/pkg/server"
	"net"
)

type Peer struct {
	conn  net.Conn
	msgCh chan server.Message
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

func NewPeer(conn net.Conn, msgCh chan server.Message) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

func (p *Peer) ReadLoop() error {
	buf := make([]byte, 1024)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			return err
		}
		msgBuf := make([]byte, n)
		copy(msgBuf, buf[:n])
		p.msgCh <- server.Message{
			Data: msgBuf,
			Peer: p,
		}
	}
}
