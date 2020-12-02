package main

import (
	"egg"
	"net/http"
)

func main() {
	r := egg.New()

	r.Get("/index", func(c *egg.Context) {
		c.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})

	v1 := r.Group("/v1")
	{
		v1.Get("/", func(c *egg.Context) {
			c.HTML(http.StatusOK, "<h1>Hello egg</h1>")
		})

		v1.Get("/hello", func(c *egg.Context) {
			// expect /hello?name=eggktutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}
	v2 := r.Group("/v2")
	{
		v2.Get("/hello/:name", func(c *egg.Context) {
			// expect /hello/eggktutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.Post("/login", func(c *egg.Context) {
			c.JSON(http.StatusOK, egg.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}

	_ = r.Run(":8080")

}
