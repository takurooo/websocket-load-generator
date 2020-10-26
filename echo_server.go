package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var addr string

var upgrader = websocket.Upgrader{}

var type2str = map[int]string{
	websocket.TextMessage:   "TextMessage",
	websocket.BinaryMessage: "BinaryMessage",
	websocket.CloseMessage:  "CloseMessage",
	websocket.PingMessage:   "PingMessage",
	websocket.PongMessage:   "PongMessage",
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	clientAddr := c.RemoteAddr()
	log.Println("client:", clientAddr.Network(), clientAddr.String())
L:
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break L
		}
		log.Printf("recv: [%s] %s", type2str[mt], message)

		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break L
		}
	}
}

func init() {
	flag.StringVar(&addr, "addr", "localhost:8000", "http service address")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", echo)
	log.Fatal(http.ListenAndServe(addr, nil))
}
