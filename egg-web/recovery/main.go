package main

import (
	"egg"
	"net/http"
)

func main() {
	r := egg.Default()
	r.Get("/", func(c *egg.Context) {
		c.String(http.StatusOK, "Hello Egg\n")
	})
	// index out of range for testing Recovery()
	r.Get("/panic", func(c *egg.Context) {
		names := []string{"DiangD"}
		c.String(http.StatusOK, names[100])
	})

	_ = r.Run(":8080")
}
