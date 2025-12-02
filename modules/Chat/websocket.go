package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	user "github.com/BVR-INNOVATION-GROUP/strike-force-backend/modules/User"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gorilla/websocket"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

// fasthttpResponseWriter wraps fasthttp.Response to implement http.ResponseWriter
type fasthttpResponseWriter struct {
	ctx *fasthttp.RequestCtx
}

func (w *fasthttpResponseWriter) Header() http.Header {
	header := make(http.Header)
	w.ctx.Response.Header.VisitAll(func(key, value []byte) {
		header.Set(string(key), string(value))
	})
	return header
}

func (w *fasthttpResponseWriter) Write(b []byte) (int, error) {
	return w.ctx.Write(b)
}

func (w *fasthttpResponseWriter) WriteHeader(statusCode int) {
	w.ctx.SetStatusCode(statusCode)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (adjust for production)
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(c *fiber.Ctx, db *gorm.DB, hub *Hub) error {
	// Get user ID from JWT
	userID := c.Locals("user_id").(uint)

	// Get group ID from query parameter
	groupIDStr := c.Query("group")
	if groupIDStr == "" {
		return c.Status(400).JSON(fiber.Map{"msg": "group parameter is required"})
	}

	groupID, err := strconv.ParseUint(groupIDStr, 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"msg": "invalid group ID"})
	}

	// Verify user has access to this group
	var group user.Group
	if err := db.Preload("Members").First(&group, uint(groupID)).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"msg": "group not found"})
	}

	// Check if user is a member or leader
	isMember := false
	if group.UserID == userID {
		isMember = true
	} else {
		for _, member := range group.Members {
			if member.ID == userID {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		return c.Status(403).JSON(fiber.Map{"msg": "you don't have access to this group"})
	}

	// Convert Fiber context to HTTP request/response for WebSocket upgrade
	httpReq, err := adaptor.ConvertRequest(c, false)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"msg": "failed to convert request: " + err.Error()})
	}

	// Create an http.ResponseWriter wrapper for fasthttp
	httpResp := &fasthttpResponseWriter{ctx: c.Context()}

	conn, err := upgrader.Upgrade(httpResp, httpReq, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return c.Status(500).JSON(fiber.Map{"msg": "failed to upgrade connection"})
	}

	client := &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, 256),
		groupID: uint(groupID),
		userID:  userID,
	}

	hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.writePump()
	go client.readPump(db, hub)

	return nil
}

// readPump pumps messages from the websocket connection to the hub
func (c *Client) readPump(db *gorm.DB, hub *Hub) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// Set read deadline
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, messageBytes, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Parse incoming message
		var incomingMessage struct {
			Body string `json:"body"`
			Type string `json:"type"`
		}

		if err := json.Unmarshal(messageBytes, &incomingMessage); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		// Create message in database
		message := Message{
			SenderID: c.userID,
			GroupID:  c.groupID,
			Body:     incomingMessage.Body,
		}

		if err := db.Create(&message).Error; err != nil {
			log.Printf("Error saving message: %v", err)
			continue
		}

		// Load sender info
		db.Preload("Sender").First(&message, message.ID)

		// Broadcast to all clients in the group
		hub.BroadcastMessage(c.groupID, message)
	}
}

// writePump pumps messages from the hub to the websocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// The hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
