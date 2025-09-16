package headers

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

var CRLF = []byte("\r\n")

var (
	ErrNoColonInHeader = "no collon in header"
	ErrMalformedHeader = "malformed header"
)

func NewHeaders() Headers {
	return Headers{}
}

type Key []byte

var KeyRegexp = regexp.MustCompile("^[A-Za-z0-9!#$%&'*+-.^_`|~]+$")

func (k *Key) isValid() bool {
	if len(*k) < 1 {
		return false
	}

	return KeyRegexp.MatchString(string(*k))
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	crlfIdx := bytes.Index(data, CRLF)
	if crlfIdx == -1 {
		return 0, false, nil
	}
	if crlfIdx == 0 {
		return 0, true, nil
	}
	headerBytes := data[:crlfIdx]

	collonIdx := bytes.Index(headerBytes, []byte(":"))
	if collonIdx == -1 {
		return 0, false, fmt.Errorf("headers parse err: %s", ErrNoColonInHeader)
	}
	if data[collonIdx-1] == ' ' {
		return 0, false, fmt.Errorf("headers parse err: %s", ErrMalformedHeader)
	}
	var key Key = bytes.TrimSpace(headerBytes[:collonIdx])
	if !key.isValid() {
		return 0, false, fmt.Errorf("headers parse err: %s", ErrMalformedHeader)
	}

	value := bytes.TrimSpace(headerBytes[collonIdx+1:])

	mapKey := strings.ToLower(string(key))
	if existingValue, ok := h[mapKey]; !ok {
		h[mapKey] = string(value)
	} else {
		h[mapKey] = existingValue + ", " + string(value)
	}

	return len(headerBytes) + len(CRLF), false, nil
}
