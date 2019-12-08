package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	targetHost := "http://localhost:8081"
	if err := run(targetHost); err != nil {
		log.Fatal(err)
	}
}

func run(targetHost string) error {
	target, err := url.Parse(targetHost)
	if err != nil {
		return err
	}
	http.Handle("/", httputil.NewSingleHostReverseProxy(target))

	return http.ListenAndServe(":8080", nil)
}
