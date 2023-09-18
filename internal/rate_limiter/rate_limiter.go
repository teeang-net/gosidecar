package main

import (
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const initialBucket uint = uint(10)

var refill uint = 1
var refillRate time.Duration = 1
var rateLimitMap = make(map[string]*Bucket)
var logger = log.New()

type Bucket = uint

func main() {
	// logger.SetFormatter(&log.JSONFormatter{})

	ticker := time.NewTicker(refillRate * time.Second)

	go func() {
		for range ticker.C {
			for k, p := range rateLimitMap {
				if *p < uint(10) {
					*p++
					logger.WithFields(log.Fields{
						"key":    k,
						"bucket": *p,
						"refill": refill,
					}).Info("New Token Added")
				}
			}
		}
	}()

	unlimitedHandler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "There's no limit to the Hello Worlds you'll get!\n")
	}

	limitedHandler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "You'll get a Hello World, but only if I feel like it.\n")
	}

	http.HandleFunc("/unlimited", unlimitedHandler)
	http.Handle("/limited", fixedWindow(http.HandlerFunc(limitedHandler)))
	logger.Fatal(http.ListenAndServe(":8080", nil))
}

func fixedWindow(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := r.RemoteAddr
		ip := strings.Split(remoteAddr, ":")[0]
		bucket, ok := rateLimitMap[ip]

		if !ok {
			log.Info("Bucket Initialized for IP: ", ip)
			b := initialBucket
			rateLimitMap[ip] = &b
			bucket = &b
		}

		if *bucket <= uint(0) {
			logger.WithFields(log.Fields{
				"ip":     ip,
				"bucket": *bucket,
			}).Warn("Rate limit exceeded")

			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, http.StatusText(http.StatusTooManyRequests))
			return
		}

		logger.WithFields(log.Fields{
			"ip":     ip,
			"bucket": *bucket,
		}).Info("Request processed")

		*bucket--
		next.ServeHTTP(w, r)
	})
}
