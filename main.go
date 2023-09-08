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
	log.SetLevel(log.DebugLevel)

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
	return &httputil.ReverseProxy{
		Transport: http.DefaultTransport,
		Director: func(r *http.Request) {
			r.URL.Scheme = target.Scheme
			r.URL.Host = target.Host
			r.Host = target.Host
		},
		ModifyResponse: func(r *http.Response) error {
			return proxyResponseMiddleware(r, target)
		},
	}
}

func proxyResponseMiddleware(r *http.Response, target *url.URL) error {
	requestLogs, requestBuf, err := readBodyBuffer(r.Request.Body, r.Request.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	r.Request.Body = io.NopCloser(bytes.NewBuffer(requestBuf))

	responseLogs, responseBuf, err := readBodyBuffer(r.Body, r.Header.Get("Content-Type"))
	if err != nil {
		return err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(responseBuf))

	var logs = ResponseLogs{
		URL:        target.ResolveReference(r.Request.URL).String(),
		StatusCode: r.StatusCode,
		Method:     r.Request.Method,
		Request:    requestLogs,
		Response:   responseLogs,
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

func readBodyBuffer(r io.ReadCloser, contentType string) (BodyLogs, []byte, error) {
	bodyLogs := BodyLogs{
		Body:        nil,
		ContentType: contentType,
	}

	buf, err := io.ReadAll(r)
	if err != nil || len(buf) == 0 {
		return bodyLogs, buf, err
	}

	if contentType == "text/html" || contentType == "application/xml" {
		bodyLogs.Body = string(buf)
		return bodyLogs, buf, nil
	}

	var body interface{}
	err = json.Unmarshal(buf, &body)

	if err != nil {
		return bodyLogs, buf, err
	}

	bodyLogs.Body = body
	return bodyLogs, buf, nil
}
