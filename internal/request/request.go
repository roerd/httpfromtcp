package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/roerd/httpfromtcp/internal/headers"
)

type RequestStatus int

const (
	requestStateInitialized RequestStatus = iota
	requestStateHeaders
	requestStateDone
)

const bufferSize = 8

type Request struct {
	RequestLine   RequestLine
	Headers       headers.Headers
	RequestStatus RequestStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := Request{RequestStatus: requestStateInitialized}

	buf := make([]byte, bufferSize)
	totalBytesRead := 0
	for request.RequestStatus != requestStateDone {
		if totalBytesRead >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[totalBytesRead:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if request.RequestStatus != requestStateDone {
					return nil, fmt.Errorf("error: reached EOF before request was fully parsed")
				}
				break
			}
			return nil, fmt.Errorf("error reading from reader: %v", err)
		}
		totalBytesRead += n

		numBytesConsumed, err := request.parse(buf[:totalBytesRead])
		if err != nil {
			return nil, err
		}

		if numBytesConsumed > 0 {
			// shift the buffer to remove the consumed bytes
			copy(buf, buf[numBytesConsumed:totalBytesRead])
			totalBytesRead -= numBytesConsumed
		}
	}

	return &request, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.RequestStatus != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed + n, err
		}
		if n == 0 {
			// not enough data to parse the next line in the request yet
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.RequestStatus {
	case requestStateInitialized:
		requestLine, numBytesConsumed, err := parseRequestLine(string(data))
		if err != nil {
			return numBytesConsumed, err
		}

		if numBytesConsumed == 0 {
			// not enough data to parse the request line yet
			return 0, nil
		}

		r.RequestLine = *requestLine
		r.RequestStatus = requestStateHeaders
		return numBytesConsumed, nil
	case requestStateHeaders:
		if r.Headers == nil {
			r.Headers = headers.NewHeaders()
		}
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return n, err
		}
		if done {
			r.RequestStatus = requestStateDone
		}
		return n, nil
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("error: unknown state")
	}
}

func parseRequestLine(request string) (*RequestLine, int, error) {
	requestLines := strings.Split(string(request), "\r\n")

	if len(requestLines) < 2 {
		return nil, 0, nil
	}

	requestLine := requestLines[0]
	numBytesConsumed := len(requestLines[0]) + len("\r\n")

	parts := strings.Split(requestLine, " ")
	if len(parts) != 3 {
		return nil, numBytesConsumed, fmt.Errorf("wrong number of parts in request line: %v", len(parts))
	}

	method := parts[0]
	if !hasOnlyCapitalLetters(method) {
		return nil, numBytesConsumed, fmt.Errorf("method contains a character that is not capital letter: %v", method)
	}

	target := parts[1]

	if parts[2] != "HTTP/1.1" {
		return nil, numBytesConsumed, fmt.Errorf("unsupported HTTP version: %v", parts[1])
	}
	version := strings.Split(parts[2], "/")[1]

	return &RequestLine{version, target, method}, numBytesConsumed, nil
}

func hasOnlyCapitalLetters(s string) bool {
	for _, r := range s {
		if !unicode.IsUpper(r) {
			return false
		}
	}
	return true
}
