package organization

import "github.com/gin-gonic/gin"

func RegisterRoutes(router gin.IRouter, handler *Handler) {
	// Organizations
	router.POST("/organizations", handler.Create)
	router.GET("/organizations", handler.List)
	router.GET("/organizations/:organization_id", handler.Get)
	router.PUT("/organizations/:organization_id", handler.Update)
	router.DELETE("/organizations/:organization_id", handler.Delete)

	// Members
	router.GET("/organizations/:organization_id/members", handler.ListMembers)
	router.PUT("/organizations/:organization_id/members/:member_id", handler.UpdateMemberRole)
	router.DELETE("/organizations/:organization_id/members/:member_id", handler.RemoveMember)
}
