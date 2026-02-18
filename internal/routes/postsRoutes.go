package routes

import "github.com/gofiber/fiber/v3"

func (r *Routes) setupPostRoutes(v1 fiber.Router) {
	posts := v1.Group("/posts")
	{
		// Static routes must come BEFORE parameterized routes
		posts.Post("", r.authMiddleware.Auth(), r.postHandler.CreatePost)
		posts.Get("", r.postHandler.GetPosts)
		posts.Get("/random", r.postHandler.GetPostsRandom)
		posts.Get("/mine", r.authMiddleware.Auth(), r.postHandler.GetMyPosts)
		posts.Post("/image", r.authMiddleware.Auth(), r.postHandler.UploadImagePosts)
		posts.Get("/username/:username", r.postHandler.GetPostsByUsername)
		posts.Get("/u/:username/:slug", r.postHandler.GetPostBySlugAndUsername)
		posts.Get("/tag/:tag", r.postHandler.GetPostsByTag)

		// Parameterized routes must come AFTER static routes
		posts.Get("/:id", r.postHandler.GetPost)
		posts.Put("/:id", r.authMiddleware.Auth(), r.postHandler.UpdatePost)
		posts.Delete("/:id", r.authMiddleware.Auth(), r.postHandler.DeletePost)

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
