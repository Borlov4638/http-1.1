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
