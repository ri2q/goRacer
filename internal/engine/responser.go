package engine

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"

	connPrep "github.com/ri2q/goRacer/internal/preparer/connection"
	h2Prep "github.com/ri2q/goRacer/internal/preparer/h2"
)

/*
trackingFramesHandler runs a read loop that processes inbound h2Prep and drives connection state:

  - SETTINGS: ACKs server settings and signals settingsReady once.

  - HEADERS/DATA: if END_STREAM is present, marks the stream finished.

  - RST_STREAM: treats the stream as finished.

  - GOAWAY: sets shutDownStatus, and once active streams drain, closes the shutDown channel.
*/

func ResponseHandler(connCfg *connPrep.ConnConfig, st *SessionState, filter string) {

	// ========= Helper Function to Handling sending GoAway() ========== //
	// Helper Function we call it when receiving any Frame with END_STREAM
	// then removing the streamIDs from the st.ActiveStreams map that added by
	// markActive that at initStreamsHandler(), and check on it if still
	// any st.ActiveStreams there if not will send GoAway() Frame, and
	// closing the connection
	markFinished := func(id, last_ID uint32) {
		st.ActiveMu.Lock()
		delete(st.ActiveStreams, id)
		empty := len(st.ActiveStreams) == 0
		st.ActiveMu.Unlock()

		if empty {
			// If no active streams and we're not expecting more, shut down gracefully
			goAwayDone(st, connCfg.TlsConn, last_ID)
		}
	}

	// =========== new Thread to Handle Received Frames ================ //
	// log.Println("Starting reading the Incoming h2Prep.")
	// resp_data := ""
	var dec = h2Prep.Decoder()
	go func() {
	POINT:
		for {
			f, err := h2Prep.ReadFrame(st.Framer)
			if err != nil {
				if err == io.EOF {
					break POINT
				} else {
					fmt.Printf("%v\n", err)
					break POINT
				}
			}
			switch f.Type {
			case h2Prep.FrameSettings:

				if !f.Ack {
					// log.Println("Got Setting, Sending ACK.")
					st.FramesMu.Lock()
					if err := h2Prep.WriteSettingAck(st.Framer); err != nil {
						fmt.Printf("Error: '%v'", err)
						os.Exit(1)
						break POINT
					}
					st.FramesMu.Unlock()
				}
			case h2Prep.FrameData:
				// b := f.Payload
				// resp_data = string(b)
				// if strings.Contains(resp_data, "After") && !strings.Contains(resp_data, "already") {
				// 	fmt.Printf("Got: %v, %s", f.StreamID, resp_data)
				// 	if f.EndStream {
				// 		markFinished(f.StreamID, f.LastID)
				// 	}
				// }
				if f.EndStream {
					markFinished(f.StreamID, f.LastID)
				}
			case h2Prep.FrameHeaders:
				headers, err := dec.DecodeFull(f.HeaderFrag)
				if err != nil {
					fmt.Printf("Error: %v", err)
				}
				for _, h := range headers {
					if h.Name == ".stattus" {
						if h.Value == filter {
							fmt.Printf("[%v]: %v", f.StreamID, filter)
						}
					}
					if f.EndStream {
						markFinished(f.StreamID, f.LastID)
					}
				}
			case h2Prep.FramePing:

				if !f.Ack {
					if err := h2Prep.WritePing(st.Framer, true, st.PingData); err != nil {
						fmt.Printf("Error %v\n", err)
						os.Exit(1)
						break POINT
					}
				}
			case h2Prep.FrameRST:
				markFinished(f.StreamID, f.LastID)
			case h2Prep.FrameGoAway:
				log.Printf(`Received GOAWAY from server: ErrCode=%v, LastStreamID=%v, Debug=%q`,
					f.Ack,
					f.LastID,
					string(f.DebugData))
				// set ShutDownReady flag, stop opening new streams
				*st.ShutDownReady = true
				time.Sleep(200 * time.Millisecond)
				st.ActiveMu.Lock()
				empty := len(st.ActiveStreams) == 0
				st.ActiveMu.Unlock()
				if empty || f.ErrCode != 0x0 {
					goAwayDone(st, connCfg.TlsConn, 0)
				}
				continue
			default:
				// print("Got frame")
			}
		}
	}()
}

var goawaySent uint32

func goAwayDone(st *SessionState, tlsConn *tls.Conn, sid uint32) {
	// only one caller should succeed
	if !atomic.CompareAndSwapUint32(&goawaySent, 0, 1) {
		return // already sent by someone else
	}
	// 	gf := http2.NewFramer(tlsConn, nil)
	// 	return gf.WriteGoAway(sid, http2.ErrCodeNo, nil)
	// log.Println("Sending GOAWAY! from below goroutine")
	if err := h2Prep.WriteGoAway(sid, tlsConn); err != nil {
	}

	time.AfterFunc(200*time.Millisecond, func() {
		select {
		case <-st.ShutDown:
			// already closed
		default:
			close(st.ShutDown)
		}
	})

}
