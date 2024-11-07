package main

import (
	"io"
	"log"
	"net/http"
)

func main() {
	log.Println("HELLO GOPHERMART!")
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Println("GET: HELLO GOPHERMART!")
		io.WriteString(w, "HELLO GOPHERMART!\n")
	}
	http.HandleFunc("/", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	//a := gophermart.New()
	//a.Run()
}
