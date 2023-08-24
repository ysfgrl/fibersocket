package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/ysfgrl/fibersocket"
)

func main() {
	app := fiber.New()
	serverSocket := fibersocket.NewSocketServer()
	serverSocket.OnConnection(func(fs *fibersocket.FiberSocket) {
		fmt.Println("--------onConnection---------")
		fmt.Println("Server ID:" + fs.ServerUUID)
		fmt.Println("Socket ID:" + fs.UUID)

		userId := fs.GetParam("id")
		fs.SetAttribute("user_id", userId)
		fs.Emit([]byte(fmt.Sprintf("Welcome to fibersocket: %s  UUID: %s", userId, fs.UUID)))
	})
	serverSocket.OnMessage(func(fs *fibersocket.FiberSocket, message []byte) {
		fmt.Println("--------New Message---------")
		fmt.Println("Server ID:" + fs.ServerUUID)
		fmt.Println("Socket ID:" + fs.UUID)
		fmt.Println("Message  :" + string(message))
	})
	serverSocket.OnError(func(fs *fibersocket.FiberSocket, err error) {
		fmt.Println("--------onError---------")
		fmt.Println("Server ID:" + fs.ServerUUID)
		fmt.Println("Socket ID:" + fs.UUID)
		fmt.Println("Error    :" + err.Error())
	})
	serverSocket.OnDisconnect(func(fs *fibersocket.FiberSocket) {
		fmt.Println("--------onDisconnect---------")
		fmt.Println("Server ID:" + fs.ServerUUID)
		fmt.Println("Socket ID:" + fs.UUID)
	})
	app.Use(cors.New())
	app.Get("/ws/:id", serverSocket.NewSocket())
	app.Post("/ws/:id", serverSocket.NewSocket())

	app.Listen(":3000")
}
