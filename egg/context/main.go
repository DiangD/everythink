package main

import (
	"egg"
	"net/http"
)

func main() {
	r := egg.New()
	r.Get("/", func(c *egg.Context) {
		c.HTML(http.StatusOK, "<h1>Hello Egg</h1>")
	})
	r.Get("/hello", func(c *egg.Context) {
		// expect /hello?name=egg
		c.String(http.StatusOK, "Welcome come to %s,dear %s", c.Path, c.Query("name"))
	})

	r.Post("/login", func(c *egg.Context) {
		c.JSON(http.StatusOK, egg.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})

	_ = r.Run(":8080")
}
