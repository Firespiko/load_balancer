package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	var port int
	var name string

	flag.IntVar(&port, "port", 8001, "Backend server port")
	flag.StringVar(&name, "name", "backend", "Backend server name")
	flag.Parse()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Response from %s\n", name)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	})

	log.Printf("%s running on :%d", name, port)

	if err := http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		nil,
	); err != nil {
		log.Fatal(err)
	}
}
