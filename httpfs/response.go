package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

var statusMessages = map[int]string{
	200: "OK",
	400: "BAD REQUEST",
	404: "NOT FOUND",
	500: "INTERNAL SERVER ERROR",
}

// Response struct represents an HTTP Response
type Response struct {
	status  int
	headers map[string]string
}

// NewResponse creates an HTTP Response struct
// given for a given connection
func NewResponse() *Response {
	init := map[string]string{
		"Connection": "close",
		"Date":       time.Now().Format(time.UnixDate),
	}
	return &Response{
		headers: init,
	}
}

// Send sends the response to the client. If the path parameter is not empty,
// Send will send the file at that given path.
func (res Response) Send(status int, data, path string) []byte {
	res.status = status
	res.headers["Content-Length"] = fmt.Sprintf("%d", len(data))

	var resp io.Reader

	// send file
	if path != "" {
		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				return res.Send(400, fmt.Sprintf("Could not find file %s", path), "")
			}
			return res.SendStatus(500)
		}
		defer f.Close()

		stats, err := f.Stat()
		if err != nil {
			res.SendStatus(500)
		}
		res.headers["Content-Length"] = fmt.Sprintf("%d", stats.Size())

		resp = io.MultiReader(
			strings.NewReader(
				fmt.Sprintf(
					"HTTP/1.0 %d %s\r\n%s\r\n",
					res.status,
					statusMessages[200],
					mapToString(res.headers),
				),
			),
			f,
		)
	} else {
		resp = strings.NewReader(fmt.Sprintf(
			"HTTP/1.0 %d %s\r\n%s\r\n%s",
			res.status,
			statusMessages[res.status],
			mapToString(res.headers),
			data,
		))
	}

	var r io.Reader
	if *verbose {
		r = io.TeeReader(resp, os.Stdout)
	} else {
		r = resp
	}

	buff := new(bytes.Buffer)
	buff.ReadFrom(r)
	return buff.Bytes()
}

// SendStatus sends a response without data
func (res Response) SendStatus(status int) []byte {
	res.status = status
	res.headers["Content-Length"] = "0"
	resp := fmt.Sprintf(
		"HTTP/1.0 %d %s\r\n%s\r\n",
		status,
		statusMessages[status],
		mapToString(res.headers),
	)
	if *verbose {
		fmt.Println(resp)
	}

	return []byte(resp)
}

func mapToString(m map[string]string) string {
	s := ""
	for k, v := range m {
		s += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	return s
}
