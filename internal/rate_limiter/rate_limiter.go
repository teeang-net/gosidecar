package rate_limiter

import (
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var Logger = log.New()

func Start() {
	// logger.SetFormatter(&log.JSONFormatter{})
	tokenBuckets := NewTokenBuckets()
	tokenBuckets.Start()

	unlimitedHandler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "There's no limit to the Hello Worlds you'll get!\n")
	}

	limitedHandler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "You'll get a Hello World, but only if I feel like it.\n")
	}

	http.HandleFunc("/unlimited", unlimitedHandler)
	http.Handle("/token-bucket", tokenBuckets.Handler((http.HandlerFunc(limitedHandler))))

	Logger.Fatal(http.ListenAndServe(":8080", nil))
}
