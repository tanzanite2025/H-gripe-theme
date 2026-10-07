package main

import (
	"log"
	"net/http"

	wheelsetlacinghandler "commerce-platform/internal/api/v1/wheelsetlacing"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(context *gin.Context) {
		context.Header("Access-Control-Allow-Origin", "http://127.0.0.1:10240")
		context.Header("Access-Control-Allow-Credentials", "true")
		context.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		context.Header("Access-Control-Allow-Headers", "Content-Type, Accept-Language, X-Locale, X-Timezone, X-Request-Timestamp, X-Request-Nonce, X-Request-Signature, X-CSRF-Token, X-Device-Fingerprint")
		if context.Request.Method == http.MethodOptions {
			context.Status(http.StatusNoContent)
			context.Abort()
			return
		}
		context.Next()
	})

	handler := wheelsetlacinghandler.NewWheelsetLacingEngineeringHTTPHandler(service.NewWheelsetLacingService())
	handler.RegisterWheelsetLacingEngineeringRoutes(router.Group("/api/v1/wheelset-lacing"))
	log.Fatal((&http.Server{Addr: "127.0.0.1:10424", Handler: router}).ListenAndServe())
}
