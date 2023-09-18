package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var bucket = 10
var refill = 1
var refillRate time.Duration = 1

func main() {
	ticker := time.NewTicker(refillRate * time.Second)

	go func() {
		for range ticker.C {
			if bucket < 10 {
				fmt.Println("New token")
				bucket++
			}

			fmt.Printf("Bucket: %d, Refresh: %d\n", bucket, refill)
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
		if bucket <= 0 {
			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, http.StatusText(http.StatusTooManyRequests))
			return
		}

		fmt.Printf("Bucket: %d, Refresh: %d\n", bucket, refill)
		bucket--
		next.ServeHTTP(w, r)
		log.Println("Request processed")
	})
}
