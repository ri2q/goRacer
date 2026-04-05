package engine

import (
	"log"
	"time"

	connPrep "github.com/ri2q/goRacer/internal/preparer/connection"
	h2Prep "github.com/ri2q/goRacer/internal/preparer/h2"
	payloadPrep "github.com/ri2q/goRacer/internal/preparer/payload"
)

func RunRacers(
	Queue []*payloadPrep.Payload,
	st *SessionState,
	connCfg *connPrep.ConnConfig,
) error {

	var i uint32
	i = 1
	for _, p := range Queue {

		//======= Prepare the Hpack Headers Buffer ========//
		hpackBuffer, err := h2Prep.HpackHandler(p)
		if err != nil {
			return err
		}
		// ========================= One Stage Request ============================= //
		// GET https://test.com
		if p.EndStream && p.Body == nil {
			// added waitGrop here to make sure the header h2Prep finished first
			// because we can't send Data Frame SID 3 before Header Frame SID 3.
			for r := i; r < p.RacerNums; r += 2 {
				st.FramesMu.Lock()
				if err := h2Prep.WriteHeaders(st.Framer, r, hpackBuffer.Bytes(), true); err != nil {
					log.Println("Error sent streamID: ", r)
				}
				st.FramesMu.Unlock()
				st.ActiveMu.Lock()
				st.ActiveStreams[r] = struct{}{}
				st.ActiveMu.Unlock()
				i += 2 // to ensure that the final ID at this loop same as the start the next stream
			}
			i += 2 // this for not use the last SID
			continue
		}

		// POST https://test.com {body}
		if p.EndStream {
			// added waitGrop here to make sure the header h2Prep finished first
			// because we can't send Data Frame SID 3 before Header Frame SID 3.
			for r := i; r < p.RacerNums; r += 2 {
				// h2Prep.Write headers
				st.FramesMu.Lock()
				if err := h2Prep.WriteHeaders(st.Framer, r, hpackBuffer.Bytes(), false); err != nil {
					log.Println("Error sent streamID: ", r, err)
				}
				st.FramesMu.Unlock()
				st.ActiveMu.Lock()
				st.ActiveStreams[r] = struct{}{}
				st.ActiveMu.Unlock()
			}

			// Here We continue sending the data h2Prep
			connCfg.DisableDelay()
			for r := i; r < p.RacerNums; r += 2 {
				st.RacersWg.Add(1)
				// Sending the Data
				go func(n uint32) {
					defer st.RacersWg.Done()
					if err := h2Prep.WriteBody(p.EndStream, p.Body, r, connCfg); err != nil {
						log.Println("Error sending data to streamID:", n, err)
					}
				}(r)
				i += 2
			}
			st.RacersWg.Wait()
			log.Println("Head Of Racers Finished:", p.RacerNums)
			i += 2
			continue
		}
		// ========================================================================= //

		// ======================= Two Stage Request =========================//
		// If endStream=false, Then the Requests will be sent with two stages
		// First stage will be the Headers with out EndStream and then after it arrive
		// it will sent approximal 35-39 DATA FRAME with EndStream with one TCP Segment
		// this (35-39) due to Maximum Segment Size: 1460
		// ----------------------------------------- //
		// Sending Headers
		for r := i; r < p.RacerNums; r += 2 {
			st.FramesMu.Lock()
			if err := h2Prep.WriteHeaders(st.Framer, r, hpackBuffer.Bytes(), p.EndStream); err != nil {
				log.Println("Error sent Header streamID: ", r, err)
			}
			st.FramesMu.Unlock()
			st.ActiveMu.Lock()
			st.ActiveStreams[r] = struct{}{}
			st.ActiveMu.Unlock()
		}

		// Sending Data h2Prep if there is a Body
		if p.Body != nil {
			connCfg.DisableDelay()
			for r := i; r < p.RacerNums; r += 2 {
				st.RacersWg.Add(1)
				// Sending the Data
				go func(n uint32) {
					defer st.RacersWg.Done()
					if err := h2Prep.WriteBody(p.EndStream, p.Body, r, connCfg); err != nil {
						log.Println("Error sending data to streamID:", n, err)
					}
				}(r)
			}
			log.Println("Body Sent.")
			st.RacersWg.Wait()
		}
		// log.Println("Head Of Racers Finished from Racers:", p.RacerNums)
		// log.Println("sleep time: ", p.HoldingTime)
		// ----------------------------------------- //
		// Sending Ping Frame To Warm Up the Connection
		connCfg.EnableDelay()
		st.RacersWg.Add(1)
		go func(hold time.Duration) {
			defer st.RacersWg.Done()
			//how much time you wanna to hold the connection before sending the end stream frames
			end := time.After(time.Millisecond * hold)
			// start := time.Now()
			log.Println("start for Ping fram")
			for {
				select {
				case <-end:
					// fmt.Println("Holding Racers for:", time.Since(start))
					return
				default:
					// this will not hold the frames, it will control how much ping frame should be send
					time.Sleep(time.Millisecond * 100)
					if err := h2Prep.WritePing(st.Framer, false, st.PingData); err != nil {
						log.Println("WritePing failed:", err)
					}
					log.Println("Pinging..")
				}
			}
		}(p.HoldingTime)
		st.RacersWg.Wait()

		// ----------------------------------------- //
		// Sending Last Data Frames With EndStream Flag to finish the Requests
		var n uint32
		st.RacersWg.Add(1)
		time.AfterFunc(time.Millisecond*0, func() {
			log.Println("Starting Sending EndStreamS")
			defer st.RacersWg.Done()
			// ----------------------------------------- //
			n, _ = FinishStream(connCfg, p.RacerNums, st, i)
		})
		st.RacersWg.Wait()
		i = n
		// ================================================================= //
	}
	return nil
}
