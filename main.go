package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

var (
	host        = flag.String("b", "0.0.0.0", "bind address")
	port        = flag.Int("p", 8080, "port to listen to")
	httpHeaders = flag.Bool("w", false, "send http headers before the data")
	format      = flag.String("f", "text", "specify the content type (text|binary|html|json|xml)")
)

func main() {
	flag.Parse()
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *host, *port))

	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	conn, err := ln.Accept()

	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}

	if flag.NArg() == 0 {
		// read from stdin -> chunked transfer if http else copy
		if *httpHeaders {
			// send http headers
			// send chunked file
			copyChunkedWithHeaders(conn, os.Stdin)
		} else {
			io.Copy(conn, os.Stdin)
		}

	} else {
		if *httpHeaders {
			copyWithHeaders(conn, flag.Args())
		} else {
			copyWithoutHeaders(conn, flag.Args())
		}

	}
}

func copyFile(w io.Writer, name string) error {
	f, err := os.Open(name)

	if err != nil {
		return err
	}
	defer f.Close()

	io.Copy(w, f)
	return nil
}

func copyChunkedWithHeaders(w io.Writer, r io.Reader) {

}

func copyWithHeaders(w io.Writer, names []string) {
	fmt.Fprint(w, "HTTP/1.1 200 OK\r\n")
	fmt.Fprint(w, "Connection: keep-alive\r\n")
	fmt.Fprint(w, "Server: xt\r\n")
	fmt.Fprintf(w, "Date: %s\r\n", time.Now().String())
	fmt.Fprintf(w, "Content-Type: %s\r\n", formatToMimeType(*format))
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", computeContentLength(names))
	copyWithoutHeaders(w, names)
}

func copyWithoutHeaders(w io.Writer, names []string) {
	for _, fname := range names {
		copyFile(w, fname)
	}
}

func computeContentLength(names []string) int64 {
	var sz int64
	for _, name := range names {
		f, _ := os.Open(name)
		st, _ := f.Stat()
		sz += st.Size()
	}

	return sz
}

func formatToMimeType(f string) string {
	switch f {
	case "text":
		return "text/plain"
	case "binary":
		return "application/octet-stream"
	case "html":
		return "text/html"
	case "json":
		return "application/json"
	case "xml":
		return "application/xml"
	default:
		return "text/plain"
	}
}
