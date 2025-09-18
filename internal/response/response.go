package response

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
)

type StatusCode int

const (
	StatusCodeOk                  StatusCode = 200
	StatusCodeBadRequest          StatusCode = 400
	StatusCodeInternalServerError StatusCode = 500
)

var ReasonStatusLineMap = map[StatusCode]string{
	StatusCodeOk:                  "OK",
	StatusCodeBadRequest:          "Bad Request",
	StatusCodeInternalServerError: "Internal Server Error",
}

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	reason := ReasonStatusLineMap[statusCode]
	statusLine := fmt.Sprintf("HTTP/1.1 %v %s\r\n", statusCode, reason)

	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	resHeaders := headers.NewHeaders()

	resHeaders["Content-Length"] = fmt.Sprint(contentLen)
	resHeaders["Connection"] = "close"
	resHeaders["Content-Type"] = "text/plain"

	return resHeaders
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for key, value := range headers {
		writeHeader := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Write([]byte(writeHeader))
		if err != nil {
			return err
		}
	}
	w.Write([]byte("\r\n"))
	return nil
}

type Writer struct {
	Writer io.Writer
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	reason := ReasonStatusLineMap[statusCode]
	statusLine := fmt.Sprintf("HTTP/1.1 %v %s\r\n", statusCode, reason)

	_, err := w.Writer.Write([]byte(statusLine))
	if err != nil {
		return err
	}

	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	for key, value := range headers {
		writeHeader := fmt.Sprintf("%s: %s\r\n", key, value)
		_, err := w.Writer.Write([]byte(writeHeader))
		if err != nil {
			return err
		}
	}
	w.Writer.Write([]byte("\r\n"))
	return nil
}

func (w *Writer) WriteBody(p []byte) (int, error) {
	bytesRead, err := w.Writer.Write(p)
	if err != nil {
		return 0, nil
	}
	return bytesRead, nil
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	chunk := fmt.Sprintf("%X\r\n%s\r\n", len(p), string(p))
	bytesWrote, err := w.Writer.Write([]byte(chunk))
	if err != nil {
		return 0, err
	}
	return bytesWrote, nil
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	endChunk := "0\r\n\r\n"
	bytesWrote, err := w.Writer.Write([]byte(endChunk))
	if err != nil {
		return 0, err
	}
	return bytesWrote, nil
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for key, value := range h {
		trailer := fmt.Sprintf("%s: %s", key, value)
		_, err := w.Writer.Write([]byte(trailer))
		if err != nil {
			return err
		}
	}
	return nil
}
