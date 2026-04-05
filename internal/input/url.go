package input

import (
	preparer "github.com/ri2q/goRacer/internal/preparer/payload"
)

// https://example.com/path?login=123 -i 12 -h key=val -c key=val -d data -j json

type urlHandler struct {
	Method         string
	Target         string
	SleepTime      int64
	CookieJar      map[string]string
	SensHeaders    map[string]string
	NonSensHeaders map[string]string
	Data           string
	Json           string
}

func NewUrlHandler() *urlHandler {
	return &urlHandler{
		Method: "",
		Data:   "",
		Json:   "",
		// Pre-allocate space for ~5-10 headers/cookies to save CPU cycles
		CookieJar:      make(map[string]string, 5),
		SensHeaders:    make(map[string]string, 5),
		NonSensHeaders: make(map[string]string, 10),
	}
}

func (u *urlHandler) Parse() (*preparer.Payload, error) {
	p := preparer.NewPayload()
	// println("u.target: ", u.Target)
	if err := p.SetTarget(u.Target); err != nil {
		return nil, err
	}
	if u.Method == "" {
		p.Method = "GET"
	} else {
		p.Method = u.Method
	}

	if u.Data != "" {
		if _, exists := u.NonSensHeaders["content-type"]; !exists {
			u.NonSensHeaders["content-type"] = "application/x-www-form-urlencoded"
		}
		if err := p.SetBody("data", u.Data); err != nil {
			return nil, err
		}
	} else if u.Json != "" {
		if _, exists := u.NonSensHeaders["content-type"]; !exists {
			u.NonSensHeaders["content-type"] = "application/json"
		}
		if err := p.SetBody("json", u.Json); err != nil {
			return nil, err
		}
	}

	if u.NonSensHeaders != nil {
		p.NonSensHeaders = u.NonSensHeaders
	}
	if u.CookieJar != nil {
		p.CookieJar = u.CookieJar
	}
	if u.SensHeaders != nil {
		p.SensitiveHeaders = u.SensHeaders
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
