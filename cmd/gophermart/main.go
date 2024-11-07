package main

import (
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"io"
	"log"
	"net/http"
)

func main() {
	config.ParseFlags()
	log.Println("RUN_ADDRESS", config.RunAddress)

	log.Println("HELLO GOPHERMART!")
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		log.Println("GET: HELLO GOPHERMART!")
		io.WriteString(w, "HELLO GOPHERMART!\n")
	}
	http.HandleFunc("/", helloHandler)
	log.Fatal(http.ListenAndServe(config.RunAddress, nil))
	//a := gophermart.New()
	//a.Run()
}
