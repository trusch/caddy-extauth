package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("---")
		r.Write(os.Stdout)
		w.Write([]byte("You are authenticated!"))
	})
	http.ListenAndServe(":8001", nil)
}
