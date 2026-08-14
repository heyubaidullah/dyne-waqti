package api

import (
	"net/http"
)

// NewRouter registers every route. Handlers under /api/v1/admin/ and the
// logout endpoint are wrapped in d.Auth.Middleware — every state-changing
// action requires a valid session, per the spec's security requirements.
// /display, /admin (the page shells), display-data, sse, and login are
// intentionally public.
func NewRouter(d *Deps) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /display", handleDisplayStub)
	mux.HandleFunc("GET /admin", handleAdminStub)
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(d.Cfg.UploadsDir))))

	mux.HandleFunc("POST /api/v1/auth/login", d.handleLogin)
	mux.Handle("POST /api/v1/auth/logout", d.Auth.Middleware(http.HandlerFunc(d.handleLogout)))

	mux.HandleFunc("GET /api/v1/display-data", d.handleDisplayData)
	mux.HandleFunc("GET /api/v1/sse", d.Broadcaster.ServeSSE)

	mux.Handle("POST /api/v1/admin/prayer-times", d.Auth.Middleware(http.HandlerFunc(d.handleUpdatePrayerTimes)))
	mux.Handle("POST /api/v1/admin/slides", d.Auth.Middleware(http.HandlerFunc(d.handleCreateSlide)))
	mux.Handle("PATCH /api/v1/admin/slides/{id}", d.Auth.Middleware(http.HandlerFunc(d.handleUpdateSlide)))
	mux.Handle("DELETE /api/v1/admin/slides/{id}", d.Auth.Middleware(http.HandlerFunc(d.handleDeleteSlide)))
	mux.Handle("POST /api/v1/admin/janazah", d.Auth.Middleware(http.HandlerFunc(d.handleJanazah)))
	mux.Handle("POST /api/v1/admin/blackout", d.Auth.Middleware(http.HandlerFunc(d.handleBlackout)))

	return mux
}
