package routers

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CorsAllowAll() gin.HandlerFunc {
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"*"}
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("authorization")

	return cors.New(corsConfig)
}
