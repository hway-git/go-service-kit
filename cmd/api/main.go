package main

import (
	"log"
	"net/http"

	"example.com/go-service-kit/internal/httpserver"
	"example.com/go-service-kit/internal/language"
	"example.com/go-service-kit/internal/startup"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	runtime, err := startup.Load("api")
	if err != nil {
		return err
	}
	defer runtime.Close()
	if err := runtime.Connect(false, true, false); err != nil {
		return err
	}
	router, err := runtime.Router()
	if err != nil {
		return err
	}

	router.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": language.T(c, "你好")})
	})

	return httpserver.Run(runtime.Addr, router)
}
