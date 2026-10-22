package routes

import (
	"github.com/gofiber/fiber/v3"
)

func (r *Routes) setupGuildRoutes(api fiber.Router) {
	guilds := api.Group("/guilds")
	{
		// Public directory.
		guilds.Get("", r.guildHandler.ListGuilds)

		// Static segments are registered before "/:slug" so a guild can never
		// shadow them (same pattern as the post routes).
		guilds.Post("", r.authMiddleware.Auth(), r.guildHandler.CreateGuild)
		guilds.Get("/me", r.authMiddleware.Auth(), r.guildHandler.GetMyGuilds)

		// Public-but-personalized: membership fields appear only when signed in.
		guilds.Get("/:slug", r.authMiddleware.OptionalAuth(), r.guildHandler.GetGuild)
		guilds.Get("/:slug/members", r.authMiddleware.OptionalAuth(), r.guildHandler.GetMembers)

		guilds.Patch("/:slug", r.authMiddleware.Auth(), r.guildHandler.UpdateGuild)
		guilds.Delete("/:slug", r.authMiddleware.Auth(), r.guildHandler.DeleteGuild)
		guilds.Post("/:slug/join", r.authMiddleware.Auth(), r.guildHandler.JoinGuild)
		guilds.Delete("/:slug/leave", r.authMiddleware.Auth(), r.guildHandler.LeaveGuild)

		// Channels: visible to whoever can see the guild; managed by owner/admin.
		guilds.Get("/:slug/channels", r.authMiddleware.OptionalAuth(), r.guildChannelHandler.ListChannels)
		guilds.Post("/:slug/channels", r.authMiddleware.Auth(), r.guildChannelHandler.CreateChannel)
		guilds.Patch("/:slug/channels/:channelId", r.authMiddleware.Auth(), r.guildChannelHandler.UpdateChannel)
		guilds.Delete("/:slug/channels/:channelId", r.authMiddleware.Auth(), r.guildChannelHandler.DeleteChannel)

		// Channel messages: members only.
		guilds.Get("/:slug/channels/:channelId/messages", r.authMiddleware.Auth(), r.guildChannelHandler.ListMessages)
		guilds.Post("/:slug/channels/:channelId/messages", r.authMiddleware.Auth(), r.guildChannelHandler.SendMessage)
		guilds.Patch("/:slug/channels/:channelId/messages/:messageId", r.authMiddleware.Auth(), r.guildChannelHandler.EditMessage)
		guilds.Delete("/:slug/channels/:channelId/messages/:messageId", r.authMiddleware.Auth(), r.guildChannelHandler.DeleteMessage)

		// Realtime events for every channel of the guild (SSE), members only.
		guilds.Get("/:slug/events", r.authMiddleware.Auth(), r.guildChannelHandler.StreamEvents)
	}
}
