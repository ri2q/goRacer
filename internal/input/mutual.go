package input

import (
	"strings"
	"time"

	preparer "github.com/ri2q/goRacer/internal/preparer/payload"
)

type Mutual struct {
	Iter       uint32
	SleepTime  int64
	EndStream  bool
	Proxy      string
	KeyLogPath string
	Filter     int
}

func NewMutual() *Mutual {
	return &Mutual{
		Iter:      25,
		SleepTime: 0,
		EndStream: false,
		Filter:    0,
	}
}

func (m *Mutual) Parse(p *preparer.Payload) {

	if m.KeyLogPath != "" {
		p.KeyLogPath = m.KeyLogPath
	}

	if m.Proxy != "" {
		// http://host:port
		if strings.Contains("socks", m.Proxy) {
			p.ProxyAddr = strings.Split(m.Proxy, "//")[1]
			p.ProxyType = "socks"
		} else if strings.Contains("//", m.Proxy) {
			p.ProxyAddr = strings.Split(m.Proxy, "//")[1]
			p.ProxyType = "http"
		} else {
			p.ProxyAddr = m.Proxy
		}

	}

	p.Filter = m.Filter
	p.RacerNums = m.Iter * 2
	p.HoldingTime = time.Duration(m.SleepTime)
	p.EndStream = m.EndStream
}
