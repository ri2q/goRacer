package preparer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/proxy"
)

// socket connects to addr optionally via a front proxy.
//   - addr: target host:port (e.g. example.com:443)
//   - proxyAddr: optional pointer to proxy address (host:port). nil => no proxy.
//   - proxyType: optional pointer to proxy type: "http" (CONNECT) or "socks"/"socks5".
//     nil or "http" will use HTTP CONNECT if proxyAddr != nil.
func tcpConnect(addr string, proxyAddr, proxyType string) (net.Conn, *net.TCPConn, syscall.RawConn, error) {
	var conn net.Conn
	var dialErr error

	// Use a dialer with timeout
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Decide proxy mode safely
	if proxyAddr != "" {

		println("proxy set", proxyAddr)
		pt := "http"
		if proxyType != "" {
			pt = strings.ToLower(proxyType)
		}

		switch pt {
		case "http":
			// Connect to proxy then issue CONNECT
			conn, dialErr = dialer.DialContext(ctx, "tcp", proxyAddr)
			if dialErr != nil {
				return nil, nil, nil, fmt.Errorf("failed to connect to proxy %s: %w", proxyAddr, dialErr)
			}

			req := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\nProxy-Connection: Keep-Alive\r\n\r\n", addr, addr)
			if _, dialErr = conn.Write([]byte(req)); dialErr != nil {
				conn.Close()
				return nil, nil, nil, fmt.Errorf("failed to send CONNECT to proxy %s: %w", proxyAddr, dialErr)
			}

			// Read proxy response until end of headers
			br := bufio.NewReader(conn)
			statusLine, err := br.ReadString('\n')
			if err != nil {
				conn.Close()
				return nil, nil, nil, fmt.Errorf("failed to read proxy response status: %w", err)
			}
			statusLine = strings.TrimSpace(statusLine)
			// Example: "HTTP/1.1 200 Connection Established"
			if !strings.Contains(statusLine, "200") {
				// read remaining headers for context
				rest, _ := br.ReadString('\n')
				conn.Close()
				return nil, nil, nil, fmt.Errorf("proxy CONNECT failed: %s -- %s", statusLine, strings.TrimSpace(rest))
			}

			// consume remaining headers until blank line
			for {
				line, err := br.ReadString('\n')
				if err != nil {
					conn.Close()
					return nil, nil, nil, fmt.Errorf("error reading proxy headers: %w", err)
				}
				if line == "\r\n" || strings.TrimSpace(line) == "" {
					break
				}
			}

			log.Printf("connected to %s via HTTP proxy %s", addr, proxyAddr)
		case "socks", "socks5":
			// Use the socks5 dialer
			dialerSocks, err := proxy.SOCKS5("tcp", proxyAddr, nil, proxy.Direct)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create SOCKS5 dialer for %s: %w", proxyAddr, err)
			}
			conn, dialErr = dialerSocks.Dial("tcp", addr)
			if dialErr != nil {
				return nil, nil, nil, fmt.Errorf("failed to dial %s via socks5 %s: %w", addr, proxyAddr, dialErr)
			}
			log.Printf("connected to %s via SOCKS5 proxy %s", addr, proxyAddr)

		default:
			return nil, nil, nil, fmt.Errorf("unsupported proxy type: %s", pt)
		}

	} else {

		// Direct connect
		conn, dialErr = dialer.DialContext(ctx, "tcp", addr)
		if dialErr != nil {
			return nil, nil, nil, fmt.Errorf("failed to dial %s: %w", addr, dialErr)
		}
		// log.Printf("connected directly to %s", addr)
	}

	// Try to get *net.TCPConn to obtain SyscallConn
	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		// Some proxy implementations return wrapper types that aren't *net.TCPConn.
		// If you need RawConn, we must fail if not convertible.
		conn.Close()
		return nil, nil, nil, fmt.Errorf("underlying connection is not *net.TCPConn; cannot obtain RawConn")
	}

	raw, err := tcp.SyscallConn()
	if err != nil {
		tcp.Close()
		return nil, nil, nil, fmt.Errorf("failed to get SyscallConn: %w", err)
	}

	return conn, tcp, raw, nil
}
