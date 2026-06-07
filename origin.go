package handler

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/middleware"
	"gitlab.com/noname-group7630520/nospam-messenger/backend/internal/ws"
)

type WSHandler struct {
	hub             *ws.Hub
	auth            AuthAPI
	msgSvc          MessageAPI
	allowedOrigins  []string
}

func NewWSHandler(hub *ws.Hub, auth AuthAPI, msgSvc MessageAPI, allowedOrigins []string) *WSHandler {
	return &WSHandler{hub: hub, auth: auth, msgSvc: msgSvc, allowedOrigins: allowedOrigins}
}

func (h *WSHandler) Serve(w http.ResponseWriter, r *http.Request) {
	sess, _, ok := middleware.AuthenticateSession(r.Context(), r, sessionCookieName, h.auth)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "authentication required")
		return
	}

	upgrader := websocket.Upgrader{
		CheckOrigin: func(req *http.Request) bool {
			return middleware.IsAllowedOrigin(req.Header.Get("Origin"), h.allowedOrigins)
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	go ws.ServeClient(context.Background(), h.hub, conn, sess.UserID, sess.Login, h.msgSvc, h.msgSvc)
}
