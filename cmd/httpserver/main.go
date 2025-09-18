package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

const port = 42069

var badRequestBody = `
	<html>
		<head>
			<title>400 Bad Request</title>
		</head>
		<body>
			<h1>Bad Request</h1>
			<p>Your request honestly kinda sucked.</p>
		</body>
	</html>
`

var internalServerErrorBody = `
	<html>
		<head>
			<title>500 Internal Server Error</title>
		</head>
		<body>
			<h1>Internal Server Error</h1>
			<p>Okay, you know what? This one is on me.</p>
		</body>
	</html>
`

var okBody = `
	<html>
		<head>
			<title>200 OK</title>
		</head>
		<body>
			<h1>Success!</h1>
			<p>Your request was an absolute banger.</p>
		</body>
	</html>
`

func proxyHandler(w *response.Writer, req *request.Request) {
	prefix := "/httpbin/"
	header := headers.NewHeaders()

	if !strings.HasPrefix(req.RequestLine.RequestTarget, prefix) {
		header["Connection"] = "close"
		header["Content-Type"] = "text/html"
		w.WriteStatusLine(response.StatusCodeBadRequest)
		header["Content-Length"] = fmt.Sprint(len(badRequestBody))
		w.WriteHeaders(header)
		w.WriteBody([]byte(badRequestBody))
		return
	}

	fullBuffer := []byte{}
	trailers := headers.NewHeaders()
	resp, err := http.Get(fmt.Sprint("https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API/Using_Fetch"))
	if err != nil {
		return
	}

	header.Add("Transfer-Encoding", "chunked")

	header.Add("Connection", "close")
	header.Add("Content-Type", "text/html")

	header.Add("Trailer", "X-Content-SHA256")
	header.Add("Trailer", "X-Content-Length")
	slog.Info("ResponseHeaders", "Headers", header)
	for {
		buffer := make([]byte, 32)
		bytesRead, err := resp.Body.Read(buffer)
		if err != nil {
			if err == io.EOF {
				w.WriteChunkedBodyDone()
			}
			break
		}
		len, err := w.WriteChunkedBody(buffer[:bytesRead])
		if err != nil {
			slog.Error("Chunced", "error", err)
			return
		}
		fullBuffer = append(fullBuffer, buffer[:bytesRead]...)
		slog.Info("Chunced", "length", len)
	}
	trailers.Add("X-Content-Length", fmt.Sprint(len(fullBuffer)))
	trailers.Add("X-Content-SHA256", fmt.Sprintf("%x", sha256.Sum256(fullBuffer)))
	w.WriteHeaders(trailers)
}

func videoHandler(w *response.Writer, req *request.Request) {
	prefix := "/video"
	header := headers.NewHeaders()

	if !strings.HasPrefix(req.RequestLine.RequestTarget, prefix) || req.RequestLine.Method != "GET" {
		header["Connection"] = "close"
		header["Content-Type"] = "text/html"
		w.WriteStatusLine(response.StatusCodeBadRequest)
		header["Content-Length"] = fmt.Sprint(len(badRequestBody))
		w.WriteHeaders(header)
		w.WriteBody([]byte(badRequestBody))
		return
	}
	header.Add("Content-Type", "video/mp4")
	w.WriteStatusLine(response.StatusCodeOk)
	w.WriteHeaders(header)

	filePath, err := filepath.Abs("./assets/vim.mp4")
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Print(err)
		panic("no vim.mp4 you dummy")
	}
	w.WriteBody(data)
}

func handler(w *response.Writer, req *request.Request) {
	headers := headers.NewHeaders()
	headers["Connection"] = "close"
	headers["Content-Type"] = "text/html"

	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteStatusLine(response.StatusCodeBadRequest)
		headers["Content-Length"] = fmt.Sprint(len(badRequestBody))
		w.WriteHeaders(headers)
		w.WriteBody([]byte(badRequestBody))
	case "/myproblem":
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		headers["Content-Length"] = fmt.Sprint(len(internalServerErrorBody))
		w.WriteHeaders(headers)
		w.WriteBody([]byte(internalServerErrorBody))
	default:
		w.WriteStatusLine(response.StatusCodeOk)
		headers["Content-Length"] = fmt.Sprint(len(okBody))
		w.WriteHeaders(headers)
		w.WriteBody([]byte(okBody))
	}
}

func main() {
	server, err := server.Serve(42069, videoHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	server.State.Store(false)
	log.Println("Server gracefully stopped")
}
