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
	r.GET("/hello", func(c *LnWeb.Context) {
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.POST("/login", func(c *LnWeb.Context) {
		c.JSON(http.StatusOK, LnWeb.H{
			"username": c.PostForm("username"),
			"password": c.PostForm("password"),
		})
	})
	r.Run(":8080")
}
