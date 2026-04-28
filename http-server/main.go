package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("error reading body: ", err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// log.Println("request: ", r.Method, r.URL.Path, r.URL.RawQuery, r.Header, r.ContentLength, r.Host, r.Body)
	log.Println("body data: ", string(body))

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello from http-server demo\n"))

	fmt.Println("==== Incoming request ==== ")
	fmt.Println("method: ", r.Method)
	fmt.Println("path: ", r.URL.Path)
	fmt.Println("query: ", r.URL.RawQuery)
	fmt.Println("headers: ", r.Header)
	fmt.Println("content length: ", r.ContentLength)
	fmt.Println("host: ", r.Host)
	fmt.Println("body: ", r.Body)
	//add some more info about the request
	fmt.Println("remote address: ", r.RemoteAddr)
	fmt.Println("request URI: ", r.RequestURI)
	fmt.Println("protocol: ", r.Proto)
	fmt.Println("user agent: ", r.UserAgent())
	fmt.Println("referer: ", r.Referer())

	fmt.Println("body data: ", string(body))

	w.WriteHeader(400)
	fmt.Fprintln(w, "\nHelo From http-server demo\n")
}

func main() {
	fmt.Println("hello http-server")

	// registers
	http.HandleFunc("/", homeHandler)

	//handle errors
	http.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Internal Server Error")
	})

	fmt.Println("Server running on localhost:8080")
	http.ListenAndServe(":8080", nil)
}
