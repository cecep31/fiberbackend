package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupBookmarkRoutes(api fiber.Router) {
	bookmarks := api.Group("/bookmarks", r.authMiddleware.Auth())
	{
		bookmarks.Post("/:post_id", r.bookmarkHandler.ToggleBookmark)
		bookmarks.Get("", r.bookmarkHandler.GetBookmarks)
		bookmarks.Patch("/:bookmark_id", r.bookmarkHandler.UpdateBookmark)
		bookmarks.Patch("/:bookmark_id/move", r.bookmarkHandler.MoveBookmark)
		bookmarks.Post("/folders", r.bookmarkHandler.CreateFolder)
		bookmarks.Get("/folders", r.bookmarkHandler.GetFolders)
		bookmarks.Patch("/folders/:folder_id", r.bookmarkHandler.UpdateFolder)
		bookmarks.Delete("/folders/:folder_id", r.bookmarkHandler.DeleteFolder)
	}
}
