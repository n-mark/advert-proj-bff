package handlers

import (
	"log/slog"
	"net/http"

	"bff-finalproj/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler exposes BFF HTTP endpoints.
type Handler struct {
	bff *service.BFF
}

// NewHandler creates a Handler.
func NewHandler(bff *service.BFF) *Handler {
	return &Handler{bff: bff}
}

// GetAdvertFull GET /api/v1/bff/adverts/:id
func (h *Handler) GetAdvertFull(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	out, err := h.bff.GetAdvertFull(c.Request.Context(), id)
	if err != nil {
		slog.Error("GetAdvertFull failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GetOrderFull GET /api/v1/bff/orders/:id
func (h *Handler) GetOrderFull(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	out, err := h.bff.GetOrderFull(c.Request.Context(), id)
	if err != nil {
		slog.Error("GetOrderFull failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// GetUserCabinet GET /api/v1/bff/users/:id/cabinet
func (h *Handler) GetUserCabinet(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	out, err := h.bff.GetUserCabinet(c.Request.Context(), id)
	if err != nil {
		slog.Error("GetUserCabinet failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, out)
}

// Health returns 200 OK.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
