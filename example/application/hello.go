package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	address := flag.String("address", ":80", "address to listen to ")

	flag.Parse()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("HELLO"))
	})

	log.Fatal(http.ListenAndServe(*address, handler))
}
