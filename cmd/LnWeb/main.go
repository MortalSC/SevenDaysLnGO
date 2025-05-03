package main

import (
	"github.com/MortalSC/SevenDaysLnGO/internal/service/LnWeb"
	"log"
	"net/http"
	"time"
)

func onlyForV2() LnWeb.HandlerFunc {
	return func(c *LnWeb.Context) {
		// Start timer
		t := time.Now()
		// if a server error occurred
		c.Fail(500, "Internal Server Error")
		// Calculate resolution time
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Request.RequestURI, time.Since(t))
	}
}

func main() {
	r := LnWeb.New()
	r.Use(LnWeb.Logger(), LnWeb.Recovery()) // global middleware
	r.GET("/", func(c *LnWeb.Context) {
		c.HTML(http.StatusOK, "static", "<h1>Hello Gee</h1>")
	})

	v2 := r.Group("/v2")
	v2.Use(onlyForV2()) // v2 group middleware
	{
		v2.GET("/hello/:name", func(c *LnWeb.Context) {
			// expect /hello/mortal
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
	}

	r.GET("/panic", func(c *LnWeb.Context) {
		sl := []int{1, 2, 3, 4, 5}
		_ = sl[100]
	})

	r.Run(":9999")
}
