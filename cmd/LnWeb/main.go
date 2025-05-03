package main

import (
	"github.com/MortalSC/SevenDaysLnGO/internal/service/LnWeb"
	"net/http"
)

func main() {
	r := LnWeb.New()
	r.GET("/", func(c *LnWeb.Context) {
		c.HTML(http.StatusOK, "<h1>Hello World</h1>")
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/", func(c *LnWeb.Context) {
			c.HTML(http.StatusOK, "<h1>Hello World</h1>")
		})

		v1.GET("/hello", func(c *LnWeb.Context) {
			// expect /hello?name=mortal
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *LnWeb.Context) {
			// expect /hello/geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *LnWeb.Context) {
			c.JSON(http.StatusOK, LnWeb.H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})

	}

	r.Run(":8080")
}
