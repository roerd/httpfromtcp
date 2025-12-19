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

func GetDefaultHeaders(contentLen int, contentType string) headers.Headers {
	return headers.Headers{
		"content-length": fmt.Sprintf("%d", contentLen),
		"connection":     "close",
		"content-type":   contentType,
	}
}

func GetNewHeaders() headers.Headers {
	return headers.NewHeaders()
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		headerLine := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(headerLine))
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}

type WriterState int

const (
	WriterStateInitial WriterState = iota
	WriterStateStatusLineWritten
	WriterStateHeadersWritten
	WriterStateBodyWritten
	WriterStateTrailersWritten
)

type Writer struct {
	writer      io.Writer
	writerState WriterState
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{
		writer:      writer,
		writerState: WriterStateInitial,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerState != WriterStateInitial {
		return fmt.Errorf("status line already written")
	}
	w.writerState = WriterStateStatusLineWritten
	return WriteStatusLine(w.writer, statusCode)
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.writerState != WriterStateStatusLineWritten {
		return fmt.Errorf("status line not written")
	}
	w.writerState = WriterStateHeadersWritten
	return WriteHeaders(w.writer, headers)
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	if w.writerState != WriterStateHeadersWritten {
		return 0, fmt.Errorf("headers not written")
	}
	w.writerState = WriterStateBodyWritten
	return w.writer.Write(p)
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.writerState != WriterStateHeadersWritten {
		return 0, fmt.Errorf("headers not written")
	}
	n, err := fmt.Fprintf(w.writer, "%X\r\n", len(p))
	if err != nil {
		return n, err
	}
	m, err := w.writer.Write(p)
	if err != nil {
		return n + m, err
	}
	l, err := w.writer.Write([]byte("\r\n"))
	return n + m + l, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	if w.writerState != WriterStateHeadersWritten {
		return 0, fmt.Errorf("headers not written")
	}
	w.writerState = WriterStateBodyWritten
	return w.writer.Write([]byte("0\r\n"))
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.writerState != WriterStateBodyWritten {
		return fmt.Errorf("body not written")
	}
	w.writerState = WriterStateTrailersWritten
	return WriteHeaders(w.writer, h)
}
