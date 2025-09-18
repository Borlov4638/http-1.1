package main

import (
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
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
	server, err := server.Serve(port, handler)
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
