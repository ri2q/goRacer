package preparer

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

func (p *Payload) SetTarget(target string) error {
	// this is due to wiered url.Parse when http[s] missing
	isHttp, err := regexp.MatchString(`^https?://`, target)
	if !isHttp {
		target = "https://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return err
	}

	host := u.Host
	// Final sanity check
	if strings.HasPrefix(host, ":") {
		return fmt.Errorf("invalid address: %s", host)
	}
	if u.Port() == "" {
		host = host + ":443"
	}
	// println(u.Host, u.Path, u.RawPath, u.RequestURI())
	if u.Path == "" {
		p.Path = "/"
	} else {
		p.Path = u.Path
	}
	if u.RawQuery != "" {
		p.Path = u.Path + "?" + u.RawQuery
	}
	p.Address = host

	return nil
}

func (p *Payload) SetBody(contentType, body string) error {
	var payload []byte
	var decoded string

	if contentType == "json" {
		decoded = strings.Trim(body, `"`)

		var js map[string]any
		if err := json.Unmarshal([]byte(decoded), &js); err != nil {
			return fmt.Errorf("Invalid JSON body: %v", err)
		}
		payload, _ = json.Marshal(js)
	} else {
		var err error
		decoded, err = url.QueryUnescape(body)
		if err != nil {
			return fmt.Errorf("Invalid form body: %v", err)
		}
		payload = []byte(decoded)
	}

	p.Body = payload
	return nil
}
