package request

import (
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

var errNeedMoreData = errors.New("need more data to process")

func (r *Request) parse(data []byte) (int, error) {
	r.ParserState = initialized

	readBytes, requestLine, err := parseRequestLine(data)
	if err != nil {
		return 0, err
	}
	if readBytes == 0 {
		return 0, nil
	}
	copy(data, data[readBytes:])

	r.ParserState = done
	r.RequestLine = *requestLine

	return readBytes, nil
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	var request Request
	buffer := make([]byte, 8)
	var total []byte

	for request.ParserState != done {
		readBytes, err := reader.Read(buffer)
		if err != nil {
			log.Fatal(err.Error())
		}

		total = append(total, buffer[:readBytes]...)
		_, err = request.parse(total)
		if err != nil {
			return nil, fmt.Errorf("error parsing data %w", err)
		}
	}

	return &request, nil
}

func parseRequestLine(dataChunk []byte) (readBytes int, requstLine *RequestLine, err error) {
	strRequest := string(dataChunk)
	var requestParts []string

	requestParts = strings.Split(strRequest, "\r\n")

	fmt.Println(requestParts)
	if strings.Contains(strRequest, "\r\n") {
		requestParts = strings.Split(strRequest, "\r\n")
	} else {
		return 0, nil, nil
	}

	requestLineParts := strings.Split(requestParts[0], " ")

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

	return len(requestParts[0]), &res, nil
}
