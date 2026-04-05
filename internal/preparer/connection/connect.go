package preparer

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
)

// ConnConfig stores all connection-level objects for TCP/TLS handling.
type ConnConfig struct {
	Conn               *net.Conn       // Base net.Conn (TCP or TLS)
	TcpConn            *net.TCPConn    // For TCP-specific socket tuning
	RawTcpSyscall      syscall.RawConn // For syscall-level socket operations
	TlsConn            *tls.Conn       // TLS-wrapped connection (if used)
	InsecureSkipVerify bool            // Allow invalid certs for testing
	KeyLogWriter       *os.File        // Enable TLS key logging (Wireshark)
	ServerName         string
}

func Connect(addr string, proxyAddr string, proxyType string, keyLogPath string) (net.Conn, *ConnConfig, error) {
	conn, tcp, rawTcpSyscall, err := tcpConnect(addr, proxyAddr, proxyType)
	if err != nil {
		return nil, nil, err
	}
	connCfg := &ConnConfig{}

	if keyLogPath != "" {
		keyLogWriter := sslKeyLogFileHandler(keyLogPath)
		defer keyLogWriter.Close()
		connCfg.Conn = &conn
		connCfg.TcpConn = tcp
		connCfg.RawTcpSyscall = rawTcpSyscall
		connCfg.InsecureSkipVerify = true
		connCfg.KeyLogWriter = keyLogWriter
		connCfg.ServerName = strings.Split(addr, ":")[0]
	} else {
		connCfg.Conn = &conn
		connCfg.TcpConn = tcp
		connCfg.RawTcpSyscall = rawTcpSyscall
		connCfg.InsecureSkipVerify = true
		connCfg.ServerName = strings.Split(addr, ":")[0]
	}

	tlsConn, err := tlsConnect(connCfg)
	if err != nil {
		return nil, nil, err
	}
	connCfg.TlsConn = tlsConn
	return conn, connCfg, nil
}

func sslKeyLogFileHandler(keyLogPath string) *os.File {

	// err := os.Setenv("SSLKEYLOGFILE", keyLogPath)
	// if err != nil {
	// 	log.Println("Warning: Couldn't Set env var!", err)
	// }
	// keyLogPath = os.Getenv("SSLKEYLOGFILE")
	// log.Println("SSLKEYLOGFILE setted: ", keyLogPath)

	var keyLogWriter *os.File
	keyLogWriter, err := os.OpenFile(keyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	println("created file: ", keyLogWriter.Name())
	if err != nil {
		log.Println("Error opening SSLKEYLOGFILE: ", err)
	}

	return keyLogWriter
}
