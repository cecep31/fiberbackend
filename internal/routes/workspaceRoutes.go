package routes

import "github.com/gofiber/fiber/v3"

func (r *Routes) setupWorkspaceRoutes(v1 fiber.Router) {
	workspaces := v1.Group("/workspaces", r.authMiddleware.Auth())
	{
		workspaces.Post("", r.workspaceHandler.CreateWorkspace)
		workspaces.Get("", r.workspaceHandler.GetAllWorkspaces)
		workspaces.Get("/me", r.workspaceHandler.GetUserWorkspaces)
		workspaces.Get("/:id", r.workspaceHandler.GetWorkspaceByID)
		workspaces.Put("/:id", r.workspaceHandler.UpdateWorkspace)
		workspaces.Delete("/:id", r.workspaceHandler.DeleteWorkspace)

		// Workspace members
		workspaces.Post("/:id/members", r.workspaceHandler.AddMember)
		workspaces.Get("/:id/members", r.workspaceHandler.GetMembers)
		workspaces.Put("/:id/members/:user_id", r.workspaceHandler.UpdateMemberRole)
		workspaces.Delete("/:id/members/:user_id", r.workspaceHandler.RemoveMember)
	}
}
