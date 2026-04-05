package preparer

import (
	"crypto/tls"

	preparer "github.com/ri2q/goRacer/internal/preparer/connection"
	"golang.org/x/net/http2"
)

func WriteHeaders(fr *http2.Framer, streamID uint32, block []byte, endStream bool) error {
	return fr.WriteHeaders(http2.HeadersFrameParam{
		StreamID:      streamID,
		BlockFragment: block,
		EndStream:     endStream,
		EndHeaders:    true,
	})
}

func WriteBody(endStream bool, body []byte, r uint32, connCfg *preparer.ConnConfig) error {
	bf := http2.NewFramer(connCfg.TlsConn, nil)
	return bf.WriteData(r, endStream, body)
}

func WritePing(fr *http2.Framer, ack bool, data [8]byte) error {
	return fr.WritePing(ack, data)
}

func WriteSettingAck(fr *http2.Framer) error {
	return fr.WriteSettingsAck()
}

func WriteGoAway(sid uint32, tlsConn *tls.Conn) error {
	gf := http2.NewFramer(tlsConn, nil)
	return gf.WriteGoAway(sid, http2.ErrCodeNo, nil)
}
