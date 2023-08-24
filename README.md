
# WebSocket wrapper for [Fiber v2](https://github.com/gofiber/fiber) 
[![Go Report Card](https://goreportcard.com/badge/github.com/ysfgrl/fibersocket)](https://goreportcard.com/report/github.com/ysfgrl/fibersocket)
[![GoDoc](https://godoc.org/github.com/ysfgrl/fibersocket?status.svg)](https://godoc.org/github.com/ysfgrl/fibersocket)
[![GitHub license](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/ysfgrl/fibersocket/blob/master/LICENSE)
### Based on [Fiber Websocket](https://github.com/gofiber/websocket) 


## ⚙️ Installation

```
go get -u github.com/ysfgrl/fibersocket
```


## ⚡️ [Examples](https://github.com/ysfgrl/fibersocket/tree/master/examples)

```go
// Single server


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
```
---

