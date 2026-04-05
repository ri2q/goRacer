package input

import (
	"bufio"
	"io"
	"net/http"
	"os"
	"strings"

	preparer "github.com/ri2q/goRacer/internal/preparer/payload"
)

type fileHandler struct {
	Raw *os.File
}

func NewFileHandler() *fileHandler {
	return &fileHandler{
		Raw: nil,
	}
}

func (f *fileHandler) Parse() (*preparer.Payload, error) {
	p := preparer.NewPayload()
	reader := bufio.NewReader(f.Raw)
	httpReq, err := http.ReadRequest(reader)
	if err != nil {
		return nil, err
	}

	target := "https://" + httpReq.Host + httpReq.RequestURI
	p.SetTarget(target)
	p.Method = httpReq.Method

	nonSens, cookies, sens := f.parseHeaders(httpReq.Header)
	if nonSens != nil {
		p.NonSensHeaders = nonSens
	}
	if cookies != nil {
		p.CookieJar = cookies
	}
	if sens != nil {
		p.SensitiveHeaders = sens
	}

	ctVal, ok := nonSens["content-type"]
	if ok {
		ctVal, _, _ = strings.Cut(ctVal, ";")
		var finalCT string
		parts := strings.Split(ctVal, "/")
		if len(parts) > 1 {
			finalCT = parts[1] // "json"
		} else {
			finalCT = parts[0] // "text" or whatever was there
		}
		ctVal = finalCT
		// println("\n", ctVal)

	}

	body, err := f.parseBody(reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		if err := p.SetBody(ctVal, *body); err != nil {
			return nil, err
		}
	}

	// println(
	// 	"\nAddress:", p.Address,
	// 	"\nbody", string(p.Body),
	// 	"\nmethod:", p.Method,
	// 	"\npath:", p.Path,
	// 	"\nRacers:", p.RacerNums,
	// 	"\nschema:", p.Schema,
	// )
	// for k, v := range p.NonSensHeaders {
	// 	println("header: ", k, "=", v)
	// }

	// for k, v := range p.SensitiveHeaders {
	// 	println("sens: ", k, "=", v)
	// }

	// for k, v := range p.CookieJar {
	// 	println("cookie: ", k, "=", v)
	// }

	return p, nil
}

// ParseHeaders return non-sensitive HTTP Headers and Sesitives HTTP Headers Respectively
// Sensitives Headers incude [cookie] and [authorization] Key
// and all that keys that have [*] prefix.
func (f *fileHandler) parseHeaders(h http.Header) (nonSens, cookies, sens map[string]string) {
	nonSens = make(map[string]string)
	cookies = make(map[string]string)
	sens = make(map[string]string)

	for key, values := range h {
		lowKey := strings.ToLower(key)
		val := strings.Join(values, "; ") // Standard HTTP join

		switch {
		case lowKey == "cookie":
			parts := strings.Split(val, ";")
			for _, p := range parts {
				kv := strings.SplitN(strings.TrimSpace(p), "=", 2)
				if len(kv) == 2 {
					cookies[kv[0]] = kv[1]
				}
			}

		case lowKey == "authorization":
			// 2. Put into Sensitive bucket
			sens[lowKey] = val

		case strings.HasPrefix(lowKey, "*"):
			lowKey = lowKey[1:]
			sens[lowKey] = val
		default:
			// 3. Everything else is Non-Sensitive
			nonSens[lowKey] = val
		}
	}

	return nonSens, cookies, sens
}

func (f *fileHandler) parseBody(r io.Reader) (*string, error) {
	var bodyPtr *string
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(b) > 2 {
		bodyStr := string(b)
		println(bodyStr)
		bodyPtr = &bodyStr
	}

	return bodyPtr, nil
}
