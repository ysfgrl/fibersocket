package fibersocket

import (
	"context"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"sync"
	"time"
)

type onMessage func(fs *FiberSocket, message []byte)
type onEvent func(fs *FiberSocket)
type onError func(fs *FiberSocket, err error)

type Message struct {
	mType   int
	message []byte
	retries int
}

type serverPool struct {
	sync.RWMutex
	serverList map[string]*SocketServer
}

type SocketServer struct {
	sync.RWMutex
	UUID         string
	socketList   map[string]*FiberSocket
	onMessage    onMessage
	onConnection onEvent
	onDisconnect onEvent
	onPing       onEvent
	onPong       onEvent
	onError      onError
}

type FiberSocket struct {
	mutex         sync.RWMutex
	websocket     *websocket.Conn
	isAlive       bool
	msgQueue      chan Message
	done          chan struct{}
	attributes    map[string]interface{}
	UUID          string
	ServerUUID    string
	GetLocals     func(key string) interface{}
	GetParam      func(key string, defaultValue ...string) string
	GetQueryParam func(key string, defaultValue ...string) string
	GetCookies    func(key string, defaultValue ...string) string
}

var sPool = serverPool{
	serverList: make(map[string]*SocketServer),
}

func NewSocketServer() *SocketServer {
	UUID := uuid.New().String()
	sPool.addServer(&SocketServer{
		socketList: make(map[string]*FiberSocket),
		UUID:       UUID,
		onMessage: func(fs *FiberSocket, data []byte) {
			fmt.Println("--------New Message---------")
			fmt.Println("Server ID:" + fs.ServerUUID)
			fmt.Println("Socket ID:" + fs.UUID)
			fmt.Println("Message  :" + string(data))
		},
		onConnection: func(fs *FiberSocket) {
			fmt.Println("--------onConnection---------")
			fmt.Println("Server ID:" + fs.ServerUUID)
			fmt.Println("Socket ID:" + fs.UUID)
		},
		onDisconnect: func(fs *FiberSocket) {
			fmt.Println("--------onDisconnect---------")
			fmt.Println("Server ID:" + fs.ServerUUID)
			fmt.Println("Socket ID:" + fs.UUID)
		},
		onError: func(fs *FiberSocket, err error) {
			fmt.Println("--------onError---------")
			fmt.Println("Server ID:" + fs.ServerUUID)
			fmt.Println("Socket ID:" + fs.UUID)
			fmt.Println("Error    :" + err.Error())
		},
		onPing: func(fs *FiberSocket) {
			fmt.Println("Socket Ping:" + fs.UUID)
		},
		onPong: func(fs *FiberSocket) {
			fmt.Println("Socket Pong:" + fs.UUID)
		},
	})
	return sPool.getServer(UUID)
}
func (ss *SocketServer) NewSocket() func(*fiber.Ctx) error {
	return websocket.New(func(c *websocket.Conn) {
		fs := &FiberSocket{
			websocket:  c,
			UUID:       uuid.New().String(),
			ServerUUID: ss.UUID,
			GetLocals: func(key string) interface{} {
				return c.Locals(key)
			},
			GetParam: func(key string, defaultValue ...string) string {
				return c.Params(key, defaultValue...)
			},
			GetQueryParam: func(key string, defaultValue ...string) string {
				return c.Query(key, defaultValue...)
			},
			GetCookies: func(key string, defaultValue ...string) string {
				return c.Cookies(key, defaultValue...)
			},
			msgQueue:   make(chan Message, 100),
			done:       make(chan struct{}, 1),
			attributes: make(map[string]interface{}),
		}
		sPool.getServer(ss.UUID).addSocket(fs)
		fs.run()
	})
}

func (sp *serverPool) addServer(fs *SocketServer) {
	sp.Lock()
	sp.serverList[fs.UUID] = fs
	sp.Unlock()
}
func (sp *serverPool) getServer(UUID string) *SocketServer {
	sp.RLock()
	ret, ok := sp.serverList[UUID]
	sp.RUnlock()
	if !ok {
		return nil
	}
	return ret
}
func GetServerById(UUID string) *SocketServer {
	return sPool.getServer(UUID)
}

func (ss *SocketServer) addSocket(fs *FiberSocket) {
	ss.Lock()
	ss.socketList[fs.UUID] = fs
	ss.Unlock()
	if ss.onConnection != nil {
		ss.onConnection(fs)
	}
}
func (ss *SocketServer) removeSocket(key string) {
	ss.Lock()
	delete(ss.socketList, key)
	ss.Unlock()
}

func (ss *SocketServer) OnMessage(callback onMessage) {
	ss.onMessage = callback
}

func (ss *SocketServer) OnError(callback onError) {
	ss.onError = callback
}
func (ss *SocketServer) OnConnection(callback onEvent) {
	ss.onConnection = callback
}

func (ss *SocketServer) OnDisconnect(callback onEvent) {
	ss.onDisconnect = callback
}

func (ss *SocketServer) OnPing(callback onEvent) {
	ss.onPing = callback
}

func (ss *SocketServer) OnPong(callback onEvent) {
	ss.onPong = callback
}

func (fs *FiberSocket) SetAttribute(key string, attribute interface{}) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.attributes[key] = attribute
}

func (fs *FiberSocket) GetAttribute(key string) interface{} {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	value, ok := fs.attributes[key]
	if ok {
		return value
	}
	return nil
}

func (fs *FiberSocket) run() {
	ctx, cancelFunc := context.WithCancel(context.Background())
	go fs.pong(ctx)
	go fs.read(ctx)
	go fs.sendQue(ctx)
	<-fs.done
	cancelFunc()
}

func (fs *FiberSocket) pong(ctx context.Context) {
	timeoutTicker := time.NewTicker(1 * time.Second)
	defer timeoutTicker.Stop()
	for {
		select {
		case <-timeoutTicker.C:
			fs.msgQueue <- Message{
				mType:   websocket.PongMessage,
				message: []byte{},
				retries: 0,
			}
		case <-ctx.Done():
			return
		}
	}
}
func (fs *FiberSocket) read(ctx context.Context) {
	timeoutTicker := time.NewTicker(10 * time.Millisecond)
	defer timeoutTicker.Stop()
	for {
		select {
		case <-timeoutTicker.C:
			if !fs.isConnection() {
				continue
			}

			fs.mutex.RLock()
			mType, msg, err := fs.websocket.ReadMessage()
			fs.mutex.RUnlock()

			if mType == websocket.PingMessage {
				sPool.getServer(fs.ServerUUID).onPing(fs)
				continue
			}

			if mType == websocket.PongMessage {
				sPool.getServer(fs.ServerUUID).onPong(fs)
				continue
			}

			if mType == websocket.CloseMessage {
				fs.disconnected(nil)
				return
			}

			if err != nil {
				fs.disconnected(err)
				return
			}
			sPool.getServer(fs.ServerUUID).onMessage(fs, msg)
		case <-ctx.Done():
			return
		}
	}
}

// Send out message queue
func (fs *FiberSocket) sendQue(ctx context.Context) {
	for {
		select {
		case message := <-fs.msgQueue:
			if !fs.isConnection() {
				if message.retries <= 5 {
					time.Sleep(20 * time.Millisecond)
					message.retries = message.retries + 1
					fs.msgQueue <- message
				}
				continue
			}
			fs.mutex.RLock()
			err := fs.websocket.WriteMessage(message.mType, message.message)
			fs.mutex.RUnlock()

			if err != nil {
				sPool.getServer(fs.ServerUUID).onDisconnect(fs)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (fs *FiberSocket) Emit(message []byte) {
	fs.msgQueue <- Message{
		mType:   websocket.TextMessage,
		message: message,
		retries: 0,
	}
}
func (fs *FiberSocket) Close() {
	fs.msgQueue <- Message{
		mType:   websocket.CloseMessage,
		message: []byte{},
		retries: 0,
	}
}

func (fs *FiberSocket) disconnected(err error) {
	sPool.getServer(fs.ServerUUID).onDisconnect(fs)
	close(fs.done)
	if err != nil {
		sPool.getServer(fs.ServerUUID).onError(fs, err)
	}
	sPool.getServer(fs.ServerUUID).removeSocket(fs.UUID)
}

func (fs *FiberSocket) isConnection() bool {
	fs.mutex.RLock()
	defer fs.mutex.RUnlock()
	return fs.websocket.Conn != nil
}
