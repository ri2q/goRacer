//go:build windows
// +build windows

package engine

import (
	"log"

	preparer "github.com/ri2q/goRacer/internal/preparer/connection"
)

/*
This function sends DATA FRAME with END_STREAM flag for all streams.
Since Windows doesn't support TCP_CORK, we use olny TCP_NODELAY.
You can find more info at https://man7.org/linux/man-pages/man7/tcp.7.html
*/
func FinishStream(c *preparer.ConnConfig, rnums uint32, st *SessionState, i uint32) (uint32, error) {
	if err := c.DisableDelay(); err != nil {
		return 0, err
	}
	for r := i; r <= rnums; r += 2 {
		st.RacersWg.Add(1)
		go func(n uint32) {
			defer st.RacersWg.Done()
			// f := http2.NewFramer(connCfg.TlsConn, connCfg.TlsConn)
			if err := st.Framer.WriteData(n, true, nil); err != nil {
				log.Println("Error sending END_STREAM for streamID: ", n)
			}
		}(r)
		i += 2
	}
	if err := c.EnableDelay(); err != nil {
		return 0, err
	}
	log.Printf("%d End of Racers Pool finished", rnums)
	i += 2
	return i, nil
}
