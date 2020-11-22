package main

import (
	engine "example/engine"
	"log"
	"net/http"
)

func main() {
	en := new(engine.Engine)
	log.Fatal(http.ListenAndServe(":8080", en))
}
