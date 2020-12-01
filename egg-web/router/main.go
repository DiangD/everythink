package main

import (
	"egg"
	"net/http"
)

func main() {
	r := egg.New()
	r.Get("/", func(c *egg.Context) {
		c.HTML(http.StatusOK, "<h1>Hello egg</h1>")
	})

	r.Get("/hello", func(c *egg.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.Get("/hello/:name/:location", func(c *egg.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Param("location"))
	})

	r.Get("/assets/*filepath", func(c *egg.Context) {
		c.JSON(http.StatusOK, egg.H{"filepath": c.Param("filepath")})
	})

	_ = r.Run(":8080")
}
