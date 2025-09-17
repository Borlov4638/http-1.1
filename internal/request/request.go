package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"log"
	"log/slog"
	"strings"
)

type ParserState string

const (
	parserStateInitialized    ParserState = "init"
	parserStateDone           ParserState = "done"
	parserStateParsingHeaders ParserState = "headers"
	parserStateParsingBody    ParserState = "body"
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
	Headers     headers.Headers
	Body        []byte
}

type RequestLine struct {
	HTTPVersion   string
	RequestTarget string
	Method        string
}

var allowedMethods = map[string]struct{}{
	"GET":  {},
	"POST": {},
}

var (
	SEPARATOR                       = "\r\n"
	CONTENT_LENGTH_HEADER           = "content-length"
	errNeedMoreData                 = errors.New("need more data to process")
	errBadContentLength             = errors.New("bad content-length")
	errNoContentLenButBodyIsPresent = errors.New("no content-length but body is presented")
)

func newRequest() Request {
	return Request{
		ParserState: parserStateInitialized,
		Headers:     headers.NewHeaders(),
	}
}

func (r *Request) hasBody() bool {
	contentLen, ok := r.Headers.GetInt(CONTENT_LENGTH_HEADER)
	slog.Info("HasBody", "ok", ok, "content-length", contentLen)
	if !ok || contentLen == 0 {
		return false
	}
	return true
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	for {
		switch r.ParserState {
		case parserStateDone:
			return read, nil
		case parserStateInitialized:
			bytesRead, requestLine, err := parseRequestLine(data[read:])
			if err != nil {
				return read, err
			}
			if bytesRead == 0 {
				return read, nil
			}
			r.RequestLine = *requestLine
			read += bytesRead
			r.ParserState = parserStateParsingHeaders
		case parserStateParsingHeaders:
			for {
				bytesRead, done, err := r.Headers.Parse(data[read:])
				if err != nil {
					return read, err
				}
				read += bytesRead
				if done {
					slog.Info("RequestHeaders", "Headers", r.Headers)
					if r.hasBody() {
						r.ParserState = parserStateParsingBody
						read = read + len(headers.CRLF)
						break
					} else {
						r.ParserState = parserStateDone
						return read + len(headers.CRLF), nil
					}
				}
				if bytesRead == 0 {
					return read, nil
				}
			}
		case parserStateParsingBody:
			fmt.Println("BODY")
			contentLength, ok := r.Headers.GetInt(CONTENT_LENGTH_HEADER)
			if !ok {
				fmt.Println(len(data[read:]) > 0)
				if len(data[read:]) > 0 {
					return read, errNoContentLenButBodyIsPresent
				}
				r.ParserState = parserStateDone
				return read, nil
			}

			remaining := min(contentLength-len(r.Body), len(data[read:]))
			slog.Info("BodyParsing", "remaining", remaining, "read", read)

			r.Body = append(r.Body, (data[read:])[:remaining]...)
			read += remaining

			if len(r.Body) > contentLength {
				return read, errBadContentLength
			} else if contentLength == len(r.Body) {
				r.ParserState = parserStateDone
			}

			return read, nil
		}
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buffer := make([]byte, 1024)
	bufferLen := 0

	for request.ParserState != parserStateDone {
		readBytes, err := reader.Read(buffer[bufferLen:])
		if err != nil || (err == io.EOF && bufferLen == 0) {
			log.Fatal(err)
		}
		bufferLen += readBytes

		slog.Info("ParsedState", "state", request.ParserState)
		consumedBytes, err := request.parse(buffer[:bufferLen])
		if err != nil {
			return nil, fmt.Errorf("error parsing data %w", err)
		}
		copy(buffer, buffer[consumedBytes:bufferLen])
		bufferLen -= consumedBytes
	}

	return &request, nil
}

func parseRequestLine(data []byte) (int, *RequestLine, error) {
	idx := bytes.Index(data, []byte(SEPARATOR))
	if idx == -1 {
		return 0, nil, nil
	}

	requstLine := string(data[:idx])

	requestLineParts := strings.Split(requstLine, " ")

	if len(requestLineParts) != 3 {
		return 0, nil, errors.New("invalid parts count")
	}

	httpVersion, _ := strings.CutPrefix(requestLineParts[2], "HTTP/")
	if httpVersion != "1.1" {
		return 0, nil, fmt.Errorf("invalid http version, presented version is %s", httpVersion)
	}

	httpMethod := requestLineParts[0]
	if strings.ToUpper(httpMethod) != httpMethod {
		return 0, nil, fmt.Errorf("invalid http method, got %s", httpMethod)
	} else if _, ok := allowedMethods[httpMethod]; !ok {
		return 0, nil, fmt.Errorf("http method is not allowed, got %s", httpMethod)
	}

	requestTarget := requestLineParts[1]
	if !strings.HasPrefix(requestTarget, "/") {
		return 0, nil, fmt.Errorf("invalid path, got %s", requestTarget)
	}

	res := RequestLine{
		HTTPVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        httpMethod,
	}

	return len(data[:idx+len(SEPARATOR)]), &res, nil
}
