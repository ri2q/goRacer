package engine

import (
	"strconv"

	connPrep "github.com/ri2q/goRacer/internal/preparer/connection"
	payloadPrep "github.com/ri2q/goRacer/internal/preparer/payload"
)

func Run(Queue []*payloadPrep.Payload) error {

	st := NewSessiontState()
	conn, connCfg, err := connPrep.Connect(
		Queue[0].Address,
		Queue[0].ProxyAddr,
		Queue[0].ProxyType,
		Queue[0].KeyLogPath,
	)
	if err != nil {
		return err
	}
	if err := StartStream(
		st,
		connCfg,
		Queue[0].MaxConcurrentStreams,
		Queue[0].InitialWindowSize,
		Queue[0].HeaderTableSize,
	); err != nil {
		return err
	}
	ResponseHandler(connCfg, st, strconv.Itoa(Queue[0].Filter))

	if err := RunRacers(Queue, st, connCfg); err != nil {
		return err
	}

	<-st.ShutDown
	conn.Close()

	return nil
}
