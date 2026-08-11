package routes

import (
	"github.com/gofiber/fiber/v3"
)

// imageBodyLimit is a per-route body pre-check for the 1MB image upload endpoint.
func imageBodyLimit(c fiber.Ctx) error {
	if c.Request().Header.ContentLength() > 1*1024*1024 {
		return fiber.ErrRequestEntityTooLarge
	}
	return c.Next()
}

func (r *Routes) setupPostRoutes(api fiber.Router) {
	posts := api.Group("/posts")
	{
		posts.Post("", r.authMiddleware.Auth(), r.postHandler.CreatePost)
		posts.Get("/random", r.postHandler.GetPostsRandom)
		posts.Get("/trending", r.postHandler.GetPostsTrending)
		posts.Get("/me", r.authMiddleware.Auth(), r.postHandler.GetMyPosts)
		posts.Get("/me/analytics", r.authMiddleware.Auth(), r.postHandler.GetMyPostsAnalytics)
		posts.Get("/me/analytics/likes-by-month", r.authMiddleware.Auth(), r.postHandler.GetMyPostsLikesByMonth)
		posts.Get("/me/:id", r.authMiddleware.Auth(), r.postHandler.GetMyPost)
		posts.Put("/me/:id", r.authMiddleware.Auth(), r.postHandler.UpdateMyPost)
		posts.Delete("/me/:id", r.authMiddleware.Auth(), r.postHandler.DeleteMyPost)
		posts.Get("/feed/for-you", r.authMiddleware.Auth(), r.postHandler.GetPostsForYou)
		posts.Post("/image", r.authMiddleware.Auth(), imageBodyLimit, r.postHandler.UploadImagePosts)
		posts.Get("/sitemap", r.postHandler.GetPostsForSitemap)
		posts.Get("/username/:username", r.postHandler.GetPostsByUsername)
		posts.Get("/u/:username/:slug", r.postHandler.GetPostBySlugAndUsername)
		posts.Get("/tag/:tag", r.postHandler.GetPostsByTag)
		posts.Get("", r.postHandler.GetPosts)
		posts.Put("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.postHandler.UpdatePost)
		posts.Delete("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.postHandler.DeletePost)
		posts.Get("/:id", r.authMiddleware.Auth(), r.authMiddleware.AuthAdmin(), r.postHandler.GetPost)

		// Comment routes
		posts.Get("/:id/comments", r.commentHandler.GetCommentsByPostID)
		posts.Post("/:id/comments", r.authMiddleware.Auth(), r.commentHandler.CreateComment)
		posts.Put("/:id/comments/:comment_id", r.authMiddleware.Auth(), r.commentHandler.UpdateComment)
		posts.Delete("/:id/comments/:comment_id", r.authMiddleware.Auth(), r.commentHandler.DeleteComment)

		// View routes
		posts.Post("/:id/view", r.authMiddleware.Auth(), r.postViewHandler.RecordView) // Only authenticated users
		posts.Get("/:id/views", r.authMiddleware.Auth(), r.postViewHandler.GetPostViews)
		posts.Get("/:id/view-stats", r.postViewHandler.GetPostViewStats)
		posts.Get("/:id/viewed", r.authMiddleware.Auth(), r.postViewHandler.CheckUserViewed)

		// Like routes
		posts.Post("/:id/like", r.authMiddleware.Auth(), r.postLikeHandler.LikePost)
		posts.Delete("/:id/like", r.authMiddleware.Auth(), r.postLikeHandler.UnlikePost)
		posts.Get("/:id/likes", r.postLikeHandler.GetPostLikes)
		posts.Get("/:id/like-stats", r.postLikeHandler.GetPostLikeStats)
		posts.Get("/:id/liked", r.authMiddleware.Auth(), r.postLikeHandler.CheckUserLiked)
	}
}
