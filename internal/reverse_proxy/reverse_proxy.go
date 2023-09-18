package reverse_proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/sirupsen/logrus"
)

func Start(targetURL string, port uint32) {
	fmt.Println("Reverse proxy listening on port", port)

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	rp := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", handleRequest(rp))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func handleRequest(rp *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(w, r)
	}
}
