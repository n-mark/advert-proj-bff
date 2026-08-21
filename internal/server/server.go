package server

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"bff-finalproj/internal/handlers"
	"bff-finalproj/internal/prometheus"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("http",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"dur_ms", time.Since(start).Milliseconds(),
		)
	}
}

// New creates a Gin router with BFF routes.
func New(h *handlers.Handler) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger())
	router.Use(prometheus.MetricsMiddleware())

	router.GET("/health", h.Health)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := router.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/bff/adverts/:id", h.GetAdvertFull)
	v1.GET("/bff/orders/:id", h.GetOrderFull)
	v1.GET("/bff/users/:id/cabinet", h.GetUserCabinet)
	
	return router
}

// Run starts the HTTP server and blocks until ctx is cancelled.
func Run(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		slog.Info("HTTP server shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("HTTP server listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// AddrFromEnv returns :$APP_PORT or :8080.
func AddrFromEnv() string {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}
