package engine

import (
	"sync"

	"golang.org/x/net/http2"
)

type SessionState struct {
	ShutDownReady *bool         // Indicates whether shutdown is allowed
	Framer        *http2.Framer // Shared framer for reading/writing frames

	ActiveStreams map[uint32]struct{} // Set of currently active stream IDs
	RacersWg      *sync.WaitGroup     // WaitGroup for racing/parallel streams
	FramesMu      *sync.Mutex
	ActiveMu      *sync.Mutex   // Guards ActiveStreams
	ShutDown      chan struct{} // Signal for closing the connection
	PingData      [8]byte       // Payload used for HTTP/2 PING frames
}

// NewRequestState allocates a fresh runtime state container.
// This should be created per connection, not per request.
func NewSessiontState() *SessionState {
	return &SessionState{
		ShutDownReady: new(bool),
		ActiveStreams: make(map[uint32]struct{}),
		RacersWg:      &sync.WaitGroup{},
		ActiveMu:      &sync.Mutex{},
		FramesMu:      &sync.Mutex{},
		ShutDown:      make(chan struct{}),
	}
}
