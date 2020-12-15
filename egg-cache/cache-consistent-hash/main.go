package main

import (
	"eggcache"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	eggcache.NewGroup("scores", 2<<10, eggcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
	// path http://localhost:8080/_eggcache/Tom
	addr := "localhost:8080"
	var peers = eggcache.NewHTTPPool(addr)
	log.Println("eggcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
