package memory

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler provides HTTP endpoints for memory and conversation management.
// SECURITY: All endpoints verify user ownership via JWT user_id.
type Handler struct {
	store Store
}

// NewHandler creates a new memory HTTP handler.
func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

// ── Conversations ──

// ListConversations handles GET /api/v1/conversations
func (h *Handler) ListConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	convs, err := h.store.ListConversations(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": convs})
}

// GetConversation handles GET /api/v1/conversations/:id
func (h *Handler) GetConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	conv, err := h.store.GetConversation(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}
	if conv == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
		return
	}

	msgs, err := h.store.GetMessages(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversation": conv, "messages": msgs})
}

// DeleteConversation handles DELETE /api/v1/conversations/:id
func (h *Handler) DeleteConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.store.DeleteConversation(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

// ── Memories ──

// ListMemories handles GET /api/v1/memories
func (h *Handler) ListMemories(c *gin.Context) {
	userID := c.GetString("user_id")

	mems, err := h.store.GetMemories(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list memories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"memories": mems})
}

// GetMemory handles GET /api/v1/memories/:id
func (h *Handler) GetMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	mem, err := h.store.GetMemory(c.Request.Context(), id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get memory"})
		return
	}
	if mem == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "memory not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"memory": mem})
}

// SearchMemories handles GET /api/v1/memories/search?q=...
func (h *Handler) SearchMemories(c *gin.Context) {
	userID := c.GetString("user_id")
	query := c.Query("q")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}
	if len(query) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too long (max 200 chars)"})
		return
	}

	mems, err := h.store.SearchMemories(c.Request.Context(), userID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search memories"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"memories": mems})
}

// DeleteMemory handles DELETE /api/v1/memories/:id
func (h *Handler) DeleteMemory(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.store.DeleteMemory(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete memory"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"deleted": true})
}
