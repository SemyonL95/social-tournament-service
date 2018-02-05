package main

import (
	"fmt"
	"net/http"
	"log"
)

func main() {

	http.HandleFunc("/", handler)
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		log.Fatal("ERROR DURING SERVING")
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "Hi there, I love %s!", r.URL.Path[1:])
}
