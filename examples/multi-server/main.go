package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/ysfgrl/fibersocket"
)

func main() {
	app := fiber.New()
	serverSocket1 := fibersocket.NewSocketServer()
	serverSocket1.OnConnection(func(fs *fibersocket.FiberSocket) {
		userId := fs.GetParam("id")
		fs.SetAttribute("user_id", userId)
		fs.Emit([]byte(fmt.Sprintf("Walcome to fibersocket: %s  UUID: %s", userId, fs.UUID)))
	})

	serverSocket2 := fibersocket.NewSocketServer()
	serverSocket2.OnConnection(func(fs *fibersocket.FiberSocket) {
		userId := fs.GetParam("id")
		fs.SetAttribute("user_id", userId)
		fs.Emit([]byte(fmt.Sprintf("Walcome to fibersocket : %s with UUID: %s", userId, fs.UUID)))
	})

	app.Use(cors.New())
	app.Get("/ws/:id", serverSocket1.NewSocket())
	app.Post("/ws/:id", serverSocket1.NewSocket())

	app.Get("/ws2/:id", serverSocket2.NewSocket())
	app.Post("/ws2/:id", serverSocket2.NewSocket())

	app.Listen(":3000")
}
