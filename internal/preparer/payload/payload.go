package preparer

import "time"

type Payload struct {
	Address              string // Target server address (host:port)Address string
	ProxyAddr            string
	ProxyType            string
	MaxConcurrentStreams uint32            // SETTINGS_MAX_CONCURRENT_STREAMS value
	InitialWindowSize    uint32            // SETTINGS_INITIAL_WINDOW_SIZE value
	HeaderTableSize      uint32            // SETTINGS_HEADER_TABLE_SIZE (HPACK)
	NonSensHeaders       map[string]string // Regular headers
	SensitiveHeaders     map[string]string // Sensitive headers (e.g., auth, cookies)
	CookieJar            map[string]string
	Filter               int
	Method               string // HTTP method (GET, POST, etc.)
	Path                 string // Request path, must be absolute
	Schema               string
	RacerNums            uint32 // Number of parallel client streams to send
	// If you set it to true, Requests will be send with END_STREAM flag
	// (i.e. the Requests will be sent immedaitly like HTTP/1.1)
	// without using HTTP/2 Feature
	EndStream   bool
	HoldingTime time.Duration // Optional delay between operations
	Body        []byte        // Optional request body
	KeyLogPath  string
}

func NewPayload() *Payload {
	return &Payload{
		Filter:               200,
		Address:              "",
		ProxyType:            "",
		ProxyAddr:            "",
		MaxConcurrentStreams: 1000,
		InitialWindowSize:    65535,
		HeaderTableSize:      4096,
		NonSensHeaders:       make(map[string]string),
		Method:               "GET",
		Path:                 "/",
		Body:                 nil,
		RacerNums:            9,
		HoldingTime:          0,
		KeyLogPath:           "",
	}
}
