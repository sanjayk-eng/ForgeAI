package invite

import "github.com/gin-gonic/gin"

// RegisterRoutes registers protected workspace invite routes
func RegisterRoutes(router gin.IRouter, handler *Handler) {
	// Protected routes (require authentication)
	router.POST("/workspaces/:workspace_id/invites", handler.CreateInvite)
	router.GET("/workspaces/:workspace_id/invites", handler.ListInvites)
	router.DELETE("/workspaces/:workspace_id/invites/:invite_id", handler.RevokeInvite)
	
	// User's pending invites (authenticated)
	router.GET("/invites/my-pending", handler.GetMyPendingInvites)
}

// RegisterPublicRoutes registers public invite routes
func RegisterPublicRoutes(router gin.IRouter, handler *Handler) {
	// Public routes (no authentication required)
	router.GET("/public/invites/:token", handler.GetInviteByToken)
	router.POST("/public/invites/:token/reject", handler.RejectInviteByToken)
	
	// Accept requires authentication
	router.POST("/invites/:token/accept", handler.AcceptInviteByToken)
}
