package service

import "fiberbackend/pkg/applog"

var (
	authLog       = applog.Component("auth")
	openRouterLog = applog.Component("openrouter")
	guildChatLog  = applog.Component("guild_chat")
	chatLog       = applog.Component("chat")
)
