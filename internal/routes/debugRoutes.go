package routes

import (
	"net/http"
	"net/http/pprof"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/adaptor"
)

func (r *Routes) setupDebugRoutes(v1 fiber.Router) {
	debug := v1.Group("/debug")
	debug.Get("/pprof/*", adaptor.HTTPHandler(http.HandlerFunc(pprof.Index)))
	debug.Get("/pprof/cmdline", adaptor.HTTPHandler(http.HandlerFunc(pprof.Cmdline)))
	debug.Get("/pprof/profile", adaptor.HTTPHandler(http.HandlerFunc(pprof.Profile)))
	debug.Get("/pprof/symbol", adaptor.HTTPHandler(http.HandlerFunc(pprof.Symbol)))
	debug.Get("/pprof/trace", adaptor.HTTPHandler(http.HandlerFunc(pprof.Trace)))
	debug.Get("/pprof/heap", adaptor.HTTPHandler(pprof.Handler("heap")))
	debug.Get("/pprof/goroutine", adaptor.HTTPHandler(pprof.Handler("goroutine")))
	debug.Get("/pprof/allocs", adaptor.HTTPHandler(pprof.Handler("allocs")))
	debug.Get("/pprof/block", adaptor.HTTPHandler(pprof.Handler("block")))
	debug.Get("/pprof/mutex", adaptor.HTTPHandler(pprof.Handler("mutex")))
}
