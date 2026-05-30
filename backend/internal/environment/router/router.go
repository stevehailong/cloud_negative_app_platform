package router

import (
	"my-cloud/internal/environment/handler"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, h *handler.EnvironmentHandler, ch *handler.ConfigHandler) {
	v1 := r.Group("/api/v1")
	{
		// 环境管理
		envGroup := v1.Group("/environments")
		{
			envGroup.GET("", h.ListEnvironments)
			envGroup.POST("", h.CreateEnvironment)
			envGroup.GET("/:id", h.GetEnvironment)
			envGroup.PUT("/:id", h.UpdateEnvironment)
			envGroup.DELETE("/:id", h.DeleteEnvironment)
		}

		// 环境模板管理
		templateGroup := v1.Group("/env-templates")
		{
			templateGroup.GET("", h.ListTemplates)
			templateGroup.POST("", h.CreateTemplate)
			templateGroup.GET("/:id", h.GetTemplate)
			templateGroup.PUT("/:id", h.UpdateTemplate)
			templateGroup.DELETE("/:id", h.DeleteTemplate)
		}

		// 应用环境绑定管理
		bindingGroup := v1.Group("/app-env-bindings")
		{
			bindingGroup.GET("", h.ListBindings)
			bindingGroup.POST("", h.CreateBinding)
			bindingGroup.GET("/:id", h.GetBinding)
			bindingGroup.PUT("/:id", h.UpdateBinding)
			bindingGroup.DELETE("/:id", h.DeleteBinding)
		}

		// ConfigMap管理
		configMapGroup := v1.Group("/config-maps")
		{
			configMapGroup.GET("", ch.ListConfigMaps)
			configMapGroup.POST("", ch.CreateConfigMap)
			configMapGroup.GET("/:id", ch.GetConfigMap)
			configMapGroup.PUT("/:id", ch.UpdateConfigMap)
			configMapGroup.DELETE("/:id", ch.DeleteConfigMap)
		}

		// Secret管理
		secretGroup := v1.Group("/secrets")
		{
			secretGroup.GET("", ch.ListSecrets)
			secretGroup.POST("", ch.CreateSecret)
			secretGroup.GET("/:id", ch.GetSecret)
			secretGroup.PUT("/:id", ch.UpdateSecret)
			secretGroup.DELETE("/:id", ch.DeleteSecret)
		}
	}
}
