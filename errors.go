package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/service"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/spam"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/ws"
)

type MessageHandler struct {
	messages MessageAPI
	hub      *ws.Hub
}

func NewMessageHandler(messages MessageAPI, hub *ws.Hub) *MessageHandler {
	return &MessageHandler{messages: messages, hub: hub}
}

func (h *MessageHandler) List(w http.ResponseWriter, r *http.Request) {
	beforeID, _ := strconv.ParseInt(r.URL.Query().Get("before_id"), 10, 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}

	msgs, err := h.messages.List(r.Context(), beforeID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"messages": toMessageList(msgs),
		"has_more": len(msgs) == limit,
	})
}

type sendRequest struct {
	Content string `json:"content"`
}

func (h *MessageHandler) Send(w http.ResponseWriter, r *http.Request) {
	userID, login, ok := SessionFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid json")
		return
	}

	m, err := h.messages.Send(r.Context(), userID, login, req.Content)
	if err != nil {
		h.mapSendError(w, err)
		return
	}
	if fresh, err := h.messages.Get(r.Context(), m.ID); err == nil {
		m = fresh
	}
	h.hub.BroadcastMessage(m)
	writeJSON(w, http.StatusCreated, map[string]any{"message": toMessageResponse(m)})
}

func (h *MessageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := SessionFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "invalid message id")
		return
	}

	m, err := h.messages.DeleteByUser(r.Context(), id, userID)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
		writeError(w, http.StatusNotFound, "not_found", "message not found")
		return
	}
	h.hub.BroadcastMessageUpdated(m)
	writeJSON(w, http.StatusOK, map[string]any{"message": toMessageResponse(m)})
}

func (h *MessageHandler) mapSendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, spam.ErrSpamRule):
		writeError(w, http.StatusBadRequest, "spam_rule", "сообщение заблокировано правилами антиспама")
	case errors.Is(err, spam.ErrRateLimited):
		writeError(w, http.StatusTooManyRequests, "rate_limited", "too many messages")
	case errors.Is(err, service.ErrMessageTooLong):
		writeError(w, http.StatusBadRequest, "message_too_long", "message exceeds 2000 characters")
	default:
		writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
