package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	var (
		dest string
		port int
	)

	flag.StringVar(&dest, "dest", "", "")
	flag.IntVar(&port, "port", 0, "")
	flag.Parse()

	if dest == "" || port == 0 {
		log.Fatal("Missing required flags: --dest and --port")
	}

	target, _ := url.Parse(dest)
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	log.Printf("Starting server on :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
