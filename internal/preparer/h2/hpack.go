package preparer

import (
	"bytes"
	"fmt"

	payloadPrep "github.com/ri2q/goRacer/internal/preparer/payload"
	"golang.org/x/net/http2/hpack"
)

func Decoder() *hpack.Decoder {
	return hpack.NewDecoder(4096, nil)
}

func HpackHandler(p *payloadPrep.Payload) (bytes.Buffer, error) {

	var hpackBuffer bytes.Buffer
	hpEnc := hpack.NewEncoder(&hpackBuffer)

	// Add required pseudo-headers for HTTP/2
	if err := hpEnc.WriteField(hpack.HeaderField{Name: ":method", Value: p.Method}); err != nil {
		return *bytes.NewBuffer(nil), err
	}
	if err := hpEnc.WriteField(hpack.HeaderField{Name: ":path", Value: p.Path}); err != nil {
		return *bytes.NewBuffer(nil), err
	}
	if err := hpEnc.WriteField(hpack.HeaderField{Name: ":scheme", Value: "https"}); err != nil {
		return *bytes.NewBuffer(nil), err
	}
	if err := hpEnc.WriteField(hpack.HeaderField{Name: ":authority", Value: p.Address}); err != nil {
		return *bytes.NewBuffer(nil), err
	}

	// Add regular HTTP headers
	if p.NonSensHeaders != nil {
		for k, v := range p.NonSensHeaders {
			if err := hpEnc.WriteField(hpack.HeaderField{Name: k, Value: v}); err != nil {
				return *bytes.NewBuffer(nil), err
			}
		}
	}

	// Add sensitive headers
	if p.SensitiveHeaders != nil {
		for k, v := range p.SensitiveHeaders {
			if err := hpEnc.WriteField(hpack.HeaderField{Name: k, Value: v, Sensitive: true}); err != nil {
				return *bytes.NewBuffer(nil), err
			}
		}
	}

	for k, v := range p.CookieJar {
		// Every entry in the jar is sent as a 'cookie' header
		cookieValue := fmt.Sprintf("%s=%s", k, v)
		err := hpEnc.WriteField(hpack.HeaderField{
			Name:      "cookie",
			Value:     cookieValue,
			Sensitive: true,
		})
		if err != nil {
			return *bytes.NewBuffer(nil), err
		}
	}

	// Return the HPACK-encoded header block
	// This buffer contains the binary representation ready for HTTP/2 HEADERS frame
	return hpackBuffer, nil
}
