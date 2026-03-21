package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/neinei960/cat/server/pkg/response"
)

func Health(c *gin.Context) {
	response.Success(c, gin.H{"status": "ok"})
}
