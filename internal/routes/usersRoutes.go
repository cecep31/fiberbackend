package routes

import "github.com/gofiber/fiber/v3"

func (r *Routes) setupUserRoutes(v1 fiber.Router) {
	users := v1.Group("/users")
	{
		// Public routes
		users.Get("/username/:username", r.userHandler.GetByUsername)
		users.Get("/:id", r.userHandler.GetByID)

		// Authenticated routes
		authUsers := users.Group("", r.authMiddleware.Auth())
		{
			authUsers.Get("/me", r.userHandler.GetMe)
			authUsers.Get("", r.authMiddleware.AuthAdmin(), r.userHandler.GetUsers)
			authUsers.Delete("/:id", r.authMiddleware.AuthAdmin(), r.userHandler.DeleteUser)

			// Follow routes
			authUsers.Post("/follow", r.userFollowHandler.FollowUser)
			authUsers.Delete("/:id/follow", r.userFollowHandler.UnfollowUser)
			authUsers.Get("/:id/follow-status", r.userFollowHandler.CheckFollowStatus)
			authUsers.Get("/:id/mutual-follows", r.userFollowHandler.GetMutualFollows)
		}

		// Follow-related public routes
		users.Get("/:id/followers", r.userFollowHandler.GetFollowers)
		users.Get("/:id/following", r.userFollowHandler.GetFollowing)
		users.Get("/:id/follow-stats", r.userFollowHandler.GetFollowStats)
	}
}
