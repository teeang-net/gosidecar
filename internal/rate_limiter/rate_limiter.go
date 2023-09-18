package main

import (
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

const initialBucket uint = uint(10)

var refill uint = 1
var refillRate time.Duration = 1
var rateLimitMap = make(map[string]*Bucket)

type Bucket = uint

func main() {
	ticker := time.NewTicker(refillRate * time.Second)

	go func() {
		for range ticker.C {
			for k, p := range rateLimitMap {
				if *p < uint(10) {
					*p++
					log.WithFields(log.Fields{
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
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fixedWindow(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		k := "hello"
		bucket, ok := rateLimitMap[k]

		if !ok {
			fmt.Println(k)
			fmt.Printf("Bucket Initialized for %s\n", k)
			b := initialBucket
			rateLimitMap[k] = &b
			bucket = &b
		}

		if *bucket <= uint(0) {
			log.WithFields(log.Fields{
				"key":    k,
				"bucket": *bucket,
			}).Warn("Rate limit exceeded")

			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, http.StatusText(http.StatusTooManyRequests))
			return
		}

		log.WithFields(log.Fields{
			"key":    k,
			"bucket": *bucket,
		}).Info("Request processed")

		*bucket--
		next.ServeHTTP(w, r)
	})
}
