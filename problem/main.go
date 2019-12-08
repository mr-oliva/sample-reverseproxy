package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"net/http/httputil"
	"os"
)

type proxy struct {
	targetHost string
	hostHeader string
	debug      bool
}

func main() {
	var debug bool
	if os.Getenv("DEBUG") == "1" {
		debug = true
	}
	p := &proxy{targetHost: "localhost:8081", hostHeader: "go.advent.2019.co.jp", debug: debug}
	if err := p.run(); err != nil {
		log.Fatal(err)
	}
}

func (p *proxy) run() error {
	http.HandleFunc("/", p.reverse)

	return http.ListenAndServe(":8080", nil)
}

func (p *proxy) reverse(w http.ResponseWriter, r *http.Request) {
	scheme := r.URL.Scheme
	if scheme == "" {
		scheme = "http"
	}
	director := func(req *http.Request) {
		req.URL.Scheme = scheme
		req.URL.Host = p.targetHost
		req.Host = p.hostHeader
	}
	if !p.debug || r.URL.Path == "/favicon.ico" {
		reverse := &httputil.ReverseProxy{Director: director}
		reverse.ServeHTTP(w, r)
		return
	}

	t := &transport{}
	trace := &httptrace.ClientTrace{
		GotConn: t.GotConn,
	}

	r = r.WithContext(httptrace.WithClientTrace(r.Context(), trace))

	reverse := &httputil.ReverseProxy{Director: director, Transport: t}
	reverse.ServeHTTP(w, r)
}

//refer to https://blog.golang.org/http-tracing
type transport struct {
	current *http.Request
}

func (t *transport) GotConn(info httptrace.GotConnInfo) {
	fmt.Printf("Connected to %v\n", t.current.URL)
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.current = req
	b, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	return http.DefaultTransport.RoundTrip(req)
}

