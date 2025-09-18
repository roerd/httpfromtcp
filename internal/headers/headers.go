package headers

import (
	"fmt"
	"regexp"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	lines := strings.Split(string(data), "\r\n")

	if len(lines) < 1 {
		return 0, false, nil
	}

	line := lines[0]

	if line == "" {
		return len("\r\n"), true, nil
	}

	key, value, found := strings.Cut(line, ":")
	if !found {
		return 0, false, fmt.Errorf("invalid header line: %q", line)
	}
	key = strings.TrimLeft(key, " \t")
	matched, _ := regexp.MatchString("^[a-zA-Z0-9!#$%&'*+.^_`|~-]+$", key)
	if !matched {
		return 0, false, fmt.Errorf("invalid header key: %q", key)
	}
	key = strings.ToLower(key)
	value = strings.TrimSpace(value)
	h[key] = value
	return len(line) + len("\r\n"), false, nil
}
