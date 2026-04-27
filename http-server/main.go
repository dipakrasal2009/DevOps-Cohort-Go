package main

import (
	"fmt"
	"io"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Println("==== Incoming request ==== ")
	fmt.Println("method: ", r.Method)
	fmt.Println("path: ", r.URL.Path)
	fmt.Println("query: ", r.URL.RawQuery)
	fmt.Println("headers: ", r.Header)
	fmt.Println("content length: ", r.ContentLength)
	fmt.Println("host: ", r.Host)
	fmt.Println("body: ", r.Body)
	// read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading body: ", err)
		return
	}
	fmt.Println("body data: ", string(body))

	w.WriteHeader(400)
	fmt.Fprintln(w, "Helo From http-server demo")
}

func main() {
	fmt.Println("hello http-server")

	// registers
	http.HandleFunc("/", homeHandler)

	fmt.Println("Server running on localhost:8080")
	http.ListenAndServe(":8080", nil)
}
