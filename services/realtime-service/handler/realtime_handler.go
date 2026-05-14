package handler

import (
	"bufio"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/runns/realtime-service/hub"
	"github.com/runns/shared/middleware"
	"github.com/runns/shared/models"
	"github.com/runns/shared/utils"
)

type RealtimeHandler struct {
	hub *hub.Hub
	rdb *redis.Client
}

func NewRealtimeHandler(h *hub.Hub, rdb *redis.Client) *RealtimeHandler {
	return &RealtimeHandler{hub: h, rdb: rdb}
}

func (h *RealtimeHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFrom(r.Context())
	role := string(middleware.RoleFrom(r.Context()))
	conn, err := upgradeWS(w, r)
	if err != nil { log.Printf("[ws] upgrade failed: %v", err); return }
	defer conn.Close()
	client := &hub.Client{ID: userID, Role: role}
	h.hub.Register(client)
	defer h.hub.Unregister(client)
	go func() {
		for msg := range client.Send { writeFrame(conn, msg) }
	}()
	for {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		payload, err := readFrame(conn)
		if err != nil { break }
		if role == string(models.RoleRider) { h.handleGPS(userID, payload) }
	}
}

func (h *RealtimeHandler) handleGPS(riderID uint, payload []byte) {
	var event models.GPSEvent
	if err := json.Unmarshal(payload, &event); err != nil { return }
	event.RiderID = riderID
	event.Timestamp = time.Now()
	ctx := context.Background()
	b, _ := json.Marshal(event)
	h.rdb.Set(ctx, fmt.Sprintf("gps:rider:%d", riderID), b, 10*time.Minute)
	h.rdb.GeoAdd(ctx, "rider:locations", &redis.GeoLocation{
		Name: fmt.Sprintf("%d", riderID), Latitude: event.Latitude, Longitude: event.Longitude,
	})
	h.rdb.Publish(ctx, "gps:events", string(b))
	msg := hub.Message{Type: hub.MsgGPSUpdate, Payload: event}
	h.hub.Broadcast("client", msg)
	h.hub.Broadcast("admin", msg)
}

func (h *RealtimeHandler) StartRedisSub(ctx context.Context) {
	pubsub := h.rdb.Subscribe(ctx, "order:events", "gps:events")
	go func() {
		for msg := range pubsub.Channel() {
			var payload interface{}
			json.Unmarshal([]byte(msg.Payload), &payload)
			h.hub.BroadcastAll(hub.Message{Type: hub.MsgOrderUpdate, Payload: payload})
		}
	}()
}

func NewRouter(h *RealtimeHandler) http.Handler {
	secret := os.Getenv("JWT_SECRET")
	auth := middleware.Authenticate(secret)
	mux := http.NewServeMux()
	mux.Handle("/ws", auth(http.HandlerFunc(h.ServeWS)))
	mux.HandleFunc("/api/realtime/location/", func(w http.ResponseWriter, r *http.Request) {
		raw := strings.TrimPrefix(r.URL.Path, "/api/realtime/location/")
		id, err := utils.ParseUint(raw, nil)
		if err != nil { utils.BadRequest(w, "invalid rider id"); return }
		val, err := h.rdb.Get(context.Background(), fmt.Sprintf("gps:rider:%d", id)).Result()
		if err != nil { utils.NotFound(w, "location not available"); return }
		var event models.GPSEvent
		json.Unmarshal([]byte(val), &event)
		utils.OK(w, "location retrieved", event)
	})
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { utils.OK(w, "ok", nil) })
	return middleware.Chain(mux, middleware.CORS, middleware.Logger)
}

func upgradeWS(w http.ResponseWriter, r *http.Request) (net.Conn, error) {
	if strings.ToLower(r.Header.Get("Upgrade")) != "websocket" {
		http.Error(w, "not a websocket", http.StatusBadRequest)
		return nil, fmt.Errorf("not ws")
	}
	key := r.Header.Get("Sec-WebSocket-Key")
	hj, ok := w.(http.Hijacker)
	if !ok { return nil, fmt.Errorf("hijack unsupported") }
	conn, buf, err := hj.Hijack()
	if err != nil { return nil, err }
	h := sha1.New()
	h.Write([]byte(key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	accept := base64.StdEncoding.EncodeToString(h.Sum(nil))
	buf.WriteString("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + accept + "\r\n\r\n")
	buf.Flush()
	return conn, nil
}

func writeFrame(conn net.Conn, data []byte) {
	frame := []byte{0x81}
	if ln := len(data); ln < 126 {
		frame = append(frame, byte(ln))
	} else {
		frame = append(frame, 126, byte(ln>>8), byte(ln))
	}
	conn.Write(append(frame, data...))
}

func readFrame(conn net.Conn) ([]byte, error) {
	reader := bufio.NewReader(conn)
	_, err := reader.ReadByte()
	if err != nil { return nil, err }
	b1, err := reader.ReadByte()
	if err != nil { return nil, err }
	masked := b1&0x80 != 0
	payloadLen := int(b1 & 0x7F)
	if payloadLen == 126 {
		b := make([]byte, 2); reader.Read(b)
		payloadLen = int(b[0])<<8 | int(b[1])
	}
	var maskKey [4]byte
	if masked { reader.Read(maskKey[:]) }
	payload := make([]byte, payloadLen)
	reader.Read(payload)
	if masked {
		for i := range payload { payload[i] ^= maskKey[i%4] }
	}
	return payload, nil
}
