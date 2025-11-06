package response

import (
	"fmt"
	"io"

	"github.com/roerd/httpfromtcp/internal/headers"
)

type StatusCode int

func (s StatusCode) String() string {
	switch s {
	case statusOK:
		return "OK"
	case statusClientError:
		return "Bad Request"
	case statusServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}

const (
	statusOK          StatusCode = 200
	statusClientError StatusCode = 400
	statusServerError StatusCode = 500
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	statusLine := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusCode.String())
	_, err := w.Write([]byte(statusLine))
	return err
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	return headers.Headers{
		"Content-Length": fmt.Sprintf("%d", contentLen),
		"Connection":     "close",
		"Content-Type":   "text/plain",
	}
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	return nil
}
