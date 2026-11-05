package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupUserRoutes(api fiber.Router) {
	users := api.Group("/users")
	{
		// Public routes
		users.Get("/username/:username", r.userHandler.GetByUsername)
		users.Get("/:id/followers", r.userFollowHandler.GetFollowers)
		users.Get("/:id/following", r.userFollowHandler.GetFollowing)
		users.Get("/:id/follow-stats", r.userFollowHandler.GetFollowStats)

		// Authenticated routes. NOTE: a Fiber Group created with handlers
		// registers them as Use() on the "/api/users" prefix, and Fiber matches
		// in registration order, so every /api/users/* route registered AFTER
		// this line runs Auth() first. Public routes must stay above it.
		authUsers := users.Group("", r.authMiddleware.Auth())
		{
			authUsers.Get("/me", r.userHandler.GetMe)
			authUsers.Get("/stats", r.authMiddleware.AuthAdmin(), r.userStatsHandler.GetStats)
			authUsers.Post("/me/image", r.userHandler.UploadAvatar)
			authUsers.Get("", r.authMiddleware.AuthAdmin(), r.userHandler.GetUsers)
			authUsers.Post("", r.authMiddleware.AuthAdmin(), r.userHandler.CreateUser)
			authUsers.Get("/:id", r.authMiddleware.AuthAdmin(), r.userHandler.GetByID)
			authUsers.Put("/:id", r.authMiddleware.AuthAdmin(), r.userHandler.UpdateUser)
			authUsers.Delete("/:id", r.authMiddleware.AuthAdmin(), r.userHandler.DeleteUser)
			authUsers.Post("/:id/restore", r.authMiddleware.AuthAdmin(), r.userHandler.RestoreUser)

			// Follow routes
			authUsers.Post("/follow", r.userFollowHandler.FollowUser)
			authUsers.Delete("/:id/follow", r.userFollowHandler.UnfollowUser)
			authUsers.Get("/:id/follow-status", r.userFollowHandler.CheckFollowStatus)
			authUsers.Get("/:id/mutual-follows", r.userFollowHandler.GetMutualFollows)
		}
	}
}
