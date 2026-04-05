package preparer

import (
	"golang.org/x/net/http2"
)

type FrameType int

const (
	FrameSettings FrameType = iota
	FrameData
	FrameHeaders
	FramePing
	FrameRST
	FrameGoAway
	FrameIgnored
)

type Frame struct {
	Type       FrameType
	StreamID   uint32
	EndStream  bool
	Payload    []byte
	HeaderFrag []byte
	DebugData  []byte
	Ack        bool
	LastID     uint32
	ErrCode    http2.ErrCode
}

func ReadFrame(fr *http2.Framer) (*Frame, error) {
	f, err := fr.ReadFrame()
	if err != nil {
		return nil, err
	}

	switch x := f.(type) {

	case *http2.SettingsFrame:
		return &Frame{Type: FrameSettings, Ack: x.IsAck()}, nil

	case *http2.DataFrame:
		return &Frame{
			Type:      FrameData,
			StreamID:  x.Header().StreamID,
			EndStream: x.StreamEnded(),
			Payload:   x.Data(),
		}, nil

	case *http2.HeadersFrame:
		return &Frame{
			Type:       FrameHeaders,
			StreamID:   x.Header().StreamID,
			EndStream:  x.StreamEnded(),
			HeaderFrag: x.HeaderBlockFragment(),
		}, nil

	case *http2.PingFrame:
		return &Frame{
			Type: FramePing,
			Ack:  x.IsAck(),
		}, nil

	case *http2.RSTStreamFrame:
		return &Frame{
			Type:     FrameRST,
			StreamID: x.Header().StreamID,
		}, nil

	case *http2.GoAwayFrame:
		return &Frame{
			Type:      FrameGoAway,
			LastID:    x.LastStreamID,
			ErrCode:   x.ErrCode,
			DebugData: x.DebugData(),
		}, nil

	default:
		// Return a non-nil frame marked as Unknown
		return &Frame{
			Type:     FrameIgnored,
			StreamID: f.Header().StreamID,
		}, nil
	}
}
