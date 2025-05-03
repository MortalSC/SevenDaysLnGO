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
		// expect /hello?name=geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
	})

	r.GET("/hello/:name", func(c *LnWeb.Context) {
		// expect /hello/geektutu
		c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
	})

	r.GET("/assets/*filepath", func(c *LnWeb.Context) {
		c.JSON(http.StatusOK, LnWeb.H{"filepath": c.Param("filepath")})
	})

	r.Run(":8080")
}
