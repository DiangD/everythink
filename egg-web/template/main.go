package main

/*
(1) render array
$ curl http://localhost:9999/date
<html>
<body>
    <p>hello, egg</p>
    <p>Date: 2019-08-17</p>
</body>
</html>
*/

/*
(2) custom render function
$ curl http://localhost:9999/students
<html>
<body>
    <p>hello, egg</p>
    <p>0: eggktutu is 20 years old</p>
    <p>1: Jack is 22 years old</p>
</body>
</html>
*/

/*
(3) serve static files
$ curl http://localhost:9999/assets/css/eggktutu.css
p {
    color: orange;
    font-weight: 700;
    font-size: 20px;
}
*/

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"egg"
)

type student struct {
	Name string
	Age  int8
}

func FormatAsDate(t time.Time) string {
	year, month, day := t.Date()
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func main() {
	r := egg.New()
	r.Use(egg.Logger())
	r.SetFuncMap(template.FuncMap{
		"FormatAsDate": FormatAsDate,
	})
	r.LoadHTMLGlob("templates/*")
	r.Static("/assets", "./static")

	stu1 := &student{Name: "DiangD", Age: 21}
	stu2 := &student{Name: "qzh", Age: 22}
	r.Get("/", func(c *egg.Context) {
		c.HTML(http.StatusOK, "css.tmpl", nil)
	})
	r.Get("/students", func(c *egg.Context) {
		c.HTML(http.StatusOK, "arr.tmpl", egg.H{
			"title":  "egg-template",
			"stuArr": [2]*student{stu1, stu2},
		})
	})

	r.Get("/date", func(c *egg.Context) {
		c.HTML(
			http.StatusOK,
			"custom_func.tmpl",
			egg.H{
				"title": "egg",
				"now":   time.Now(),
			},
		)
	})

	_ = r.Run(":8080")
}
