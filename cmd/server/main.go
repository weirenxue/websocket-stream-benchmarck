package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
	"websocket-steam-benchmark/util"

	"github.com/c2h5oh/datasize"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func init() {
	rand.Seed(time.Nanosecond.Nanoseconds())
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var dummyMessageSize datasize.ByteSize

func main() {
	// Load config from config.toml
	config, err := util.LoadConfig("config")
	if err != nil {
		log.Fatal("loadconfig:", err)
	}

	// Convert size from string to uint64.
	if err := dummyMessageSize.UnmarshalText([]byte(config.Server.DummyMessageSize)); err != nil {
		log.Fatal("unmarshal datasize:", err)
	}

	// Output the configs.
	fmt.Printf("size per dummy message: %s\n", dummyMessageSize.String())
	fmt.Printf("dummy message per second: %v\n", int(time.Second/config.Server.DummyMessageDuration))

	// Serve websocket services.
	server := NewServer(&config)
	server.Start("0.0.0.0:" + config.Server.Port)
}

type Server struct {
	config *util.Config
	chs    sync.Map
	router *gin.Engine
}

func NewServer(config *util.Config) *Server {
	server := &Server{
		config: config,
		router: gin.Default(),
	}

	server.router.GET("/ws", server.wsHandler)

	return server
}

func (server *Server) Start(address string) {
	server.router.Run(address)
}

func (server *Server) wsHandler(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// All messages to the client should be sent to this channel.
	ch := make(chan []byte, 1)
	server.chs.Store(conn, ch)

	go server.writeMessageLoop(conn, ch)
	go server.dummyEmitLoop(conn, server.config.Server.DummyMessageDuration, dummyMessageSize.Bytes())
	go server.readMessageLoop(conn)
}

func (server *Server) writeMessageLoop(conn *websocket.Conn, ch <-chan []byte) {
	for response := range ch {
		conn.WriteMessage(websocket.TextMessage, response)
	}
}

func (server *Server) readMessageLoop(conn *websocket.Conn) {
	// All messages to the client should be sent to this channel.
	value, _ := server.chs.Load(conn)
	ch, _ := value.(chan []byte)

	// Variable that calculates how many
	// request are received from the client.
	var totalRequest int

	defer func() {
		if r := recover(); r != nil {
			// Read message from the closed connection
			// will throw panic. Close and delete the
			// corresponding connection and channel.
			server.chs.Delete(conn)
			conn.Close()
			close(ch)
			log.Printf("total request %d\n", totalRequest)
		}
	}()

	// Echoes any message from the client.
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			continue
		}
		ch <- message
		totalRequest += 1
	}
}

func (server *Server) dummyEmitLoop(conn *websocket.Conn, d time.Duration, dummySize uint64) {
	var totalDummyBytes datasize.ByteSize
	defer func() {
		if r := recover(); r != nil {
			// The channel is closed because the
			// connection is closed. Write to the
			// closed channel will throw panic.
			log.Printf("total dummy message: %s\n", totalDummyBytes.String())
			return
		}
	}()

	// All messages to the client should be sent to this channel.
	value, _ := server.chs.Load(conn)
	ch, _ := value.(chan []byte)
	dummy := []byte(util.RandomString(dummySize))

	for {
		time.Sleep(d)
		ch <- dummy
		totalDummyBytes += datasize.ByteSize(dummySize)
	}
}
