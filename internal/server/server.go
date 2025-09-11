package server

import (
	"fmt"
	"net/http"
)

func myHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World!\n")
}

func Start() {
	http.HandleFunc("/", myHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error: ", err)
	}
}
