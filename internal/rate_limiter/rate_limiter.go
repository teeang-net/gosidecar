package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	unlimitedHandler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "There's no limit to the Hello Worlds you'll get!\n")
	}

	limitedHandler := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "You'll get a Hello World, but only if I feel like it.\n")
	}

	http.HandleFunc("/unlimited", unlimitedHandler)
	http.HandleFunc("/limited", limitedHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
