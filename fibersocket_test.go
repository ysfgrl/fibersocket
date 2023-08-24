package fibersocket

import (
	"fmt"
	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"testing"
)

func TestHelloRoute(t *testing.T) {

	app := fiber.New()
	serverSocket := NewSocketServer()
	serverSocket.OnMessage(func(fs *FiberSocket, message []byte) {
		fmt.Println("--------New Message---------")
		fmt.Println("Server ID:" + fs.ServerUUID)
		fmt.Println("Socket ID:" + fs.UUID)
		fmt.Println("Message  :" + string(message))
		fs.Emit(message)
	})
	app.Use(cors.New())
	app.Get("/ws/:id", serverSocket.NewSocket())
	app.Post("/ws/:id", serverSocket.NewSocket())
	go func() {
		_ = app.Listen(":3000")
	}()

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:3000/ws/123", nil)
	if err != nil {
		t.Fatalf("%v", err)
	}
	defer ws.Close()

	// Send message to server, read response and check to see if it's what we expect.
	for i := 0; i < 10; i++ {
		if err := ws.WriteMessage(websocket.TextMessage, []byte("hello")); err != nil {
			t.Fatalf("%v", err)
		}
		_, p, err := ws.ReadMessage()
		if err != nil {
			t.Fatalf("%v", err)
		}
		if string(p) != "hello" {
			t.Fatalf("bad message")
		}
	}
}
