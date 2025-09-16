package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
)

type ParserState int

const (
	initialized ParserState = iota
	done
)

type Request struct {
	RequestLine RequestLine
	ParserState ParserState
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
	SEPARATOR       = "\r\n"
	errNeedMoreData = errors.New("need more data to process")
)

func newRequest() Request {
	return Request{
		ParserState: initialized,
	}
}

func (r *Request) parse(data []byte) (int, error) {
	read := 0

	for {
		switch r.ParserState {
		case done:
			return read, nil
		case initialized:
			bytesRead, requestLine, err := parseRequestLine(data[read:])
			if err != nil {
				return read, err
			}
			if bytesRead == 0 {
				return read, nil
			}
			r.RequestLine = *requestLine
			read += bytesRead
			r.ParserState = done
		}
	}
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	buffer := make([]byte, 1024)
	bufferLen := 0

	for request.ParserState != done {
		readBytes, err := reader.Read(buffer[bufferLen:])
		if err != nil {
			log.Fatal(err.Error())
		}
		bufferLen += readBytes

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
		fmt.Println(httpMethod)
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
