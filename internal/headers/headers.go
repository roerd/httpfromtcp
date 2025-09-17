package headers

import (
	"fmt"
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
	if key == "" || strings.Contains(key, " ") || strings.Contains(key, "\t") {
		return 0, false, fmt.Errorf("invalid header key: %q", key)
	}
	value = strings.TrimSpace(value)
	h[key] = value
	return len(line) + len("\r\n"), false, nil
}
