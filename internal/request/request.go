package request

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLineStr := strings.Split(string(request), "\r\n")[0]
	requestLine, err := parseRequestLine(requestLineStr)
	if err != nil {
		return nil, err
	}

	return &Request{*requestLine}, nil
}

func parseRequestLine(requestLine string) (*RequestLine, error) {
	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, fmt.Errorf("Wrong number of parts in request line: %v", len(parts))
	}

	method := parts[0]
	if !hasOnlyCapitalLetters(method) {
		return nil, fmt.Errorf("Method contains a character that is not capital letter: %v", method)
	}
	
	target := parts[1]

	if parts[2] != "HTTP/1.1" {
		return nil, fmt.Errorf("Unsupported HTTP version: %v", parts[1])
	}
	version := strings.Split(parts[2], "/")[1]

	return &RequestLine{version, target, method}, nil
}

func hasOnlyCapitalLetters(s string) bool {
    for _, r := range s {
        if !unicode.IsUpper(r) {
            return false
        }
    }
    return true
}
