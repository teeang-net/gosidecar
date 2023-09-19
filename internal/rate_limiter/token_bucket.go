package rate_limiter

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

type Bucket = uint

type TokenBuckets struct {
	Buckets map[string]*Bucket
}

func NewTokenBuckets() *TokenBuckets {
	return &TokenBuckets{
		Buckets: make(map[string]*Bucket),
	}
}

func (rl *TokenBuckets) Start() {
	ticker := time.NewTicker(refillRate * time.Second)

	go func() {
		for range ticker.C {
			for k, p := range rl.Buckets {
				if *p < uint(10) {
					*p++
					Logger.WithFields(log.Fields{
						"key":    k,
						"bucket": *p,
						"refill": refill,
					}).Info("New Token Added")
				}
			}
		}
	}()
}

func (rl *TokenBuckets) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remoteAddr := r.RemoteAddr
		ip := strings.Split(remoteAddr, ":")[0]
		bucket, ok := rl.Buckets[ip]

		if !ok {
			Logger.Info("Bucket Initialized for IP: ", ip)
			b := initialBucket
			rl.Buckets[ip] = &b
			bucket = &b
		}

		if *bucket <= uint(0) {
			Logger.WithFields(log.Fields{
				"ip":     ip,
				"bucket": *bucket,
			}).Warn("Rate limit exceeded")

			w.WriteHeader(http.StatusTooManyRequests)
			io.WriteString(w, http.StatusText(http.StatusTooManyRequests))
			return
		}

		Logger.WithFields(log.Fields{
			"ip":     ip,
			"bucket": *bucket,
		}).Info("Request processed")

		*bucket--
		next.ServeHTTP(w, r)
	})
}
