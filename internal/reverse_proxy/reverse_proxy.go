package reverse_proxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func Start(targetURL string, port uint32) {
	fmt.Println("Reverse proxy listening on port", port)

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatal(err)
	}

	rp := httputil.NewSingleHostReverseProxy(target)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		rp.ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
