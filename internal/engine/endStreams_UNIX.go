//go:build linux
// +build linux

package engine

import (
	"log"

	preparer "github.com/ri2q/goRacer/internal/preparer/connection"
	"golang.org/x/net/http2"
)

/*
z
This function is to ensure all DATA FRAME with END_STREAM flag are sent with single TCP segment.
you can find more information about TCP_NODELAY and TCP_CORK in the following link:
https://man7.org/linux/man-pages/man7/tcp.7.html
*/
func FinishStream(connCfg *preparer.ConnConfig, rnums uint32, st *SessionState, i uint32) (uint32, error) {
	// log.Println("Using TCP_NODELAY WITH TCP_CORK..")
	if err := connCfg.Cork(); err != nil {
		return 0, err
	}

	connCfg.DisableDelay()
	for r := i; r <= rnums; r += 2 {
		st.RacersWg.Add(1)
		go func(n uint32) {
			defer st.RacersWg.Done()
			f := http2.NewFramer(connCfg.TlsConn, connCfg.TlsConn)
			// log.Println("Sending END_STREAM FOR SID: ", n)
			if err := f.WriteData(n, true, nil); err != nil {
				log.Println("Error sending END_STREAM for streamID: ", r)
			}
		}(r)
		i += 2
	}
	// log.Printf("%d End of Racers Pool finished", rnums/2)
	if err := connCfg.UnCork(); err != nil {
		return 0, err
	}
	// connCfg.TcpConn.SetNoDelay(true)
	i += 2
	return i, nil
}
