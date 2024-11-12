package main

import (
	"fmt"
	"net/http"
)

func main() {
	addr := ":8080"
	h := GetApp()

	fmt.Println("start server at", addr)
	http.ListenAndServe(addr, h)
}
