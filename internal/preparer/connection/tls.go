package preparer

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"
)

func tlsConnect(c *ConnConfig) (*tls.Conn, error) {
	tlsCfg := &tls.Config{
		NextProtos:         []string{"h2"},
		InsecureSkipVerify: c.InsecureSkipVerify,
		ServerName:         c.ServerName,
	}

	if c.KeyLogWriter != nil {
		println("set log")
		tlsCfg.KeyLogWriter = c.KeyLogWriter
	}

	tlsConn := tls.Client(c.TcpConn, tlsCfg)

	// Set a deadline for the handshake so it doesn't hang forever
	c.TcpConn.SetDeadline(time.Now().Add(5 * time.Second))
	defer c.TcpConn.SetDeadline(time.Time{}) // Reset deadline

	if err := tlsConn.Handshake(); err != nil {
		return nil, fmt.Errorf("TLS handshake failed: %w", err)
	}

	state := tlsConn.ConnectionState()
	if state.NegotiatedProtocol != "h2" {
		tlsConn.Close() // Clean up!
		return nil, fmt.Errorf("ALPN failed: got %q, want h2", state.NegotiatedProtocol)
	}

	log.Println("TLS negotiated ALPN:", state.NegotiatedProtocol)
	return tlsConn, nil
}
