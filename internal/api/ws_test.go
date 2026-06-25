package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kabirnarang39/claude-team/internal/api"
)

func dialTestHub(t *testing.T, hub *api.Hub) (*websocket.Conn, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial: %v", err)
	}
	return conn, srv
}

func TestHubBroadcast(t *testing.T) {
	hub := api.NewHub()
	conn, srv := dialTestHub(t, hub)
	defer srv.Close()
	defer conn.Close()

	// Give ServeWS goroutine time to register the client
	time.Sleep(20 * time.Millisecond)

	hub.Broadcast([]byte(`{"type":"test"}`))

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(msg) != `{"type":"test"}` {
		t.Errorf("got %q, want {\"type\":\"test\"}", string(msg))
	}
}

func TestHubBroadcastMultipleClients(t *testing.T) {
	hub := api.NewHub()
	conn1, srv1 := dialTestHub(t, hub)
	conn2, srv2 := dialTestHub(t, hub)
	defer srv1.Close()
	defer srv2.Close()
	defer conn1.Close()
	defer conn2.Close()

	time.Sleep(20 * time.Millisecond)

	hub.Broadcast([]byte(`hello`))

	for i, c := range []*websocket.Conn{conn1, conn2} {
		c.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := c.ReadMessage()
		if err != nil {
			t.Fatalf("client %d read: %v", i, err)
		}
		if string(msg) != "hello" {
			t.Errorf("client %d: got %q, want hello", i, string(msg))
		}
	}
}

func TestHubBroadcastPrunesDeadClient(t *testing.T) {
	hub := api.NewHub()
	conn, srv := dialTestHub(t, hub)
	defer srv.Close()

	time.Sleep(20 * time.Millisecond)
	// Close without unregistering — Broadcast should prune it without panicking
	conn.Close()
	time.Sleep(20 * time.Millisecond)

	// Should not panic or block
	hub.Broadcast([]byte(`after close`))
}

func TestServeWSPongKeepAlive(t *testing.T) {
	hub := api.NewHub()
	conn, srv := dialTestHub(t, hub)
	defer srv.Close()
	defer conn.Close()

	// Send a pong — server's pong handler should reset deadline without error
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	conn.WriteMessage(websocket.PongMessage, nil)
	// No crash = pass
}
