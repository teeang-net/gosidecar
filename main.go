package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
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

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	target, _ := url.Parse(dest)
	proxy := configureReverseProxy(target)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	log.WithFields(log.Fields{
		"port": port,
		"dest": dest,
	}).Info("Starting server")

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func configureReverseProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(r *http.Request) {
		r.URL.Scheme = target.Scheme
		r.URL.Host = target.Host
		r.Host = target.Host
	}

	proxy.ModifyResponse = func(r *http.Response) error {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		var bodyStr = ""
		var requestBody map[string]interface{}

		r.Body = io.NopCloser(bytes.NewReader(buf))

		// Convert HTML and XML responses to strings
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xml") {
			bodyStr = string(buf)
			requestBody = nil
		} else {
			bodyStr = ""
			err = json.Unmarshal(buf, &requestBody)
		}

		if err != nil {
			return err
		}

		log.WithFields(log.Fields{
			"body":         bodyStr,
			"Content-Type": contentType,
			"method":       r.Request.Method,
			"status":       r.Status,
			"url":          target.ResolveReference(r.Request.URL).String(),
		}).Infoln()

		return nil
	}

	return proxy
}

// Reads the provided io.Reader and unmarshals it into a map[string]interface{}.
// It returns the unmarshalled map, the original request body as a byte slice, and any errors encountered.
func UnmarshallReader(r io.Reader) (map[string]interface{}, []byte, error) {
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, body, err
	}

	var requestBody map[string]interface{}
	err = json.Unmarshal(body, &requestBody)

	if err != nil {
		return nil, body, err
	}

	return requestBody, body, nil
}
