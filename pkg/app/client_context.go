package app

import "github.com/gin-gonic/gin"

type ClientContext struct {
	GinContext *gin.Context
}

func NewContext(ginContext *gin.Context) ClientContext {
	return ClientContext{GinContext: ginContext}
}
