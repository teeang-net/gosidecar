package main

import (
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

type BodyLogs struct {
	Body        interface{}
	ContentType string
}

type ResponseLogs struct {
	URL        string
	StatusCode int
	Method     string
	Request    BodyLogs
	Response   BodyLogs
}

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
		return proxyResponseMiddleware(r, target)
	}

	return proxy
}
func proxyResponseMiddleware(r *http.Response, target *url.URL) error {
	requestContent := r.Request.Header.Get("Content-Type")
	responseContent := r.Header.Get("Content-Type")

	buf, err := io.ReadAll(r.Request.Body)
	if err != nil {
		return fmt.Errorf("error reading request body: %w", err)
	}
	requestStr, requestJSON, err := readBodyBuffer(buf, requestContent)
	if err != nil {
		return fmt.Errorf("error parsing request body: %w", err)
	}

	buf, err = io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}
	responseStr, responseJSON, err := readBodyBuffer(buf, responseContent)
	if err != nil {
		return fmt.Errorf("error parsing response body: %w", err)
	}

	var logs = ResponseLogs{
		URL:        target.ResolveReference(r.Request.URL).String(),
		StatusCode: r.StatusCode,
		Method:     r.Request.Method,
		Request: BodyLogs{
			Body:        requestStr,
			ContentType: r.Request.Header.Get("Content-Type"),
		},
		Response: BodyLogs{
			Body:        responseStr,
			ContentType: r.Header.Get("Content-Type"),
		},
	}

	if requestJSON != nil {
		logs.Request.Body = requestJSON
	}

	if responseJSON != nil {
		logs.Request.Body = responseJSON
	}

	log.WithFields(log.Fields{
		"url":        logs.URL,
		"method":     logs.Method,
		"statusCode": logs.StatusCode,
		"response":   logs.Response,
		"request":    logs.Request,
	}).Infoln()

	return nil
}

func readBodyBuffer(buf []byte, contentType string) (string, map[string]interface{}, error) {
	var bodyStr = ""
	var requestBody map[string]interface{}
	var err error

	if strings.Contains(contentType, "text/html") || strings.Contains(contentType, "application/xml") {
		bodyStr = string(buf)
		requestBody = nil
	} else {
		bodyStr = ""
		err = json.Unmarshal(buf, &requestBody)
	}

	if err != nil {
		return "", nil, err
	}

	return bodyStr, requestBody, nil
}
