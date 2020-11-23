package main

import (
	"din"
	"fmt"
	"net/http"
)

func main() {
	r := din.New()
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "URL.Path = %q\n", req.URL.Path)
	})

	r.Get("/hello", func(w http.ResponseWriter, req *http.Request) {
		for k, v := range req.Header {
			fmt.Fprintf(w, "Header[%q] = %q\n", k, v)
		}
	})
	r.Get("/test", func(w http.ResponseWriter, req *http.Request) {
		fmt.Fprintf(w, "cookie:%s\n", req.Cookies())
	})

	err := r.Run(":8080")
	if err != nil {
		panic(err)
	}
}
