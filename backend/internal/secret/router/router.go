package router

import (
	"my-cloud/internal/secret/handler"

	"github.com/gin-gonic/gin"
)

func SetupRouter(r *gin.Engine, h *handler.SecretHandler) {
	v1 := r.Group("/api/v1")
	{
		// 密钥管理
		secretGroup := v1.Group("/app-secrets")
		{
			secretGroup.GET("", h.ListSecrets)
			secretGroup.POST("", h.CreateSecret)
			secretGroup.GET("/:id", h.GetSecret)
			secretGroup.PUT("/:id", h.UpdateSecret)
			secretGroup.DELETE("/:id", h.DeleteSecret)
		}

		// 按应用查询密钥
		v1.GET("/app-secrets/apps/:appId", h.GetSecretsByApp)
	}
}
