package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("INFO: handle /hello " + req.Method)
		if _, err := w.Write([]byte("Hello world!")); err != nil {
			fmt.Println("ERR:", err)
		}
	})
	fmt.Println("INFO: start server at *:" + port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("ERR:", err)
		os.Exit(1)
	}
}
