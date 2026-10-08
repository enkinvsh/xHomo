package tun

import (
	"net"

	mips "github.com/metacubex/mipstack"
	"github.com/metacubex/sing/common"
	M "github.com/metacubex/sing/common/metadata"
)

func (s *Mipstack) forwardTCP(request *mips.TCPForwarderRequest) {
	flow := request.Flow()
	conn, err := request.Accept(s.ctx)
	if err != nil {
		return
	}
	metadata := M.Metadata{
		Source:      M.SocksaddrFromNetIP(flow.Source),
		Destination: M.SocksaddrFromNetIP(flow.Destination),
	}
	var handlerConn net.Conn = conn
	if common.LowMemory && s.mtu > 0 {
		handlerConn = &mipsLowMemoryConn{TCPConn: conn, mtu: int(s.mtu)}
	}
	go func() {
		if err := s.handler.NewConnection(s.ctx, handlerConn, metadata); err != nil {
			_ = conn.SetLinger(0)
			_ = conn.Close()
		}
	}()
}

// mipsLowMemoryConn reports the TUN MTU to relays, which size their copy
// buffers from it; otherwise each direction takes a 16 KB buffer and holds it
// for as long as the connection sits idle.
type mipsLowMemoryConn struct {
	*mips.TCPConn
	mtu int
}

func (c *mipsLowMemoryConn) ReaderMTU() int { return c.mtu }

func (c *mipsLowMemoryConn) WriterMTU() int { return c.mtu }

func (c *mipsLowMemoryConn) Upstream() any { return c.TCPConn }
