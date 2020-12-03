package main

import (
	"egg"
	"log"
	"net/http"
	"time"
)

func main() {
	r := egg.New()
	r.Use(egg.Logger())
	r.Get("/", func(c *egg.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.Get("/hello/:name", func(c *egg.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	_ = r.Run(":8080")
}

func onlyForV2() egg.HandlerFunc {
	return func(c *egg.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}
