package service

import "fiberbackend/pkg/applog"

var (
	authLog       = applog.Component("auth")
	openRouterLog = applog.Component("openrouter")
)
