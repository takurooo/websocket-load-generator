package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Arg struct {
	url         string
	duration    int
	connections int
	length      int
	unit        string
}

var arg Arg

type Config struct {
	wg       *sync.WaitGroup
	id       int
	done     chan struct{}
	url      string
	duration time.Duration
	length   int
}

func clinet(config Config) {
	defer config.wg.Done()

	c, _, err := websocket.DefaultDialer.Dial(config.url, nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}
	defer c.Close()

	ticker := time.NewTicker(config.duration)
	defer ticker.Stop()

	message := strings.Repeat("0", config.length)

	for {
		select {
		case <-ticker.C:
			// log.Println(config.id, t.String())
			start := time.Now()
			if err := c.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				log.Println("write:", err)
				return
			}
			_, _, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			elapsedTime := time.Now().Sub(start)
			log.Println("id:", config.id, "rtt", elapsedTime.String())

		case <-config.done:
			closeMessage := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprint(config.id))
			if err := c.WriteMessage(websocket.CloseMessage, closeMessage); err != nil {
				log.Println("write close:", err)
				return
			}
			log.Println("client close:", config.id)
			return
		}
	}
}

func init() {
	flag.StringVar(&arg.url, "a", "ws://localhost:8000/", "http service address")
	flag.IntVar(&arg.duration, "d", 1, "duration")
	flag.IntVar(&arg.connections, "c", 1, "connections")
	flag.StringVar(&arg.unit, "u", "sec", "duration unit(msec or sec)")
	flag.IntVar(&arg.length, "l", 1, "message length")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	u, err := url.Parse(arg.url)
	if err != nil {
		log.Fatalln("url parse error", err)
	}
	if u.Scheme != "ws" && u.Scheme != "wss" {
		log.Fatalln("invalid scheme:", u.Scheme)
	}
	log.Printf("connecting to %s", u.String())

	length := arg.length
	if length <= 0 {
		log.Fatalln("invalid length:", length)
	}

	connections := arg.connections
	if connections <= 0 {
		log.Fatalln("invalid connections:", connections)
	}

	duration := arg.duration
	if duration <= 0 {
		log.Fatalln("invalid duration:", duration)
	}

	var unit time.Duration
	switch arg.unit {
	case "msec":
		unit = time.Millisecond
	case "sec":
		unit = time.Second
	default:
		log.Fatalln("invalid unit:", unit)
	}

	// -------------------------
	// create clients
	// -------------------------
	wg := &sync.WaitGroup{}
	done := make(chan struct{})

	for i := 0; i < connections; i++ {
		wg.Add(1)
		go clinet(Config{
			wg:       wg,
			id:       i,
			done:     done,
			url:      u.String(),
			duration: time.Duration(duration) * unit,
			length:   arg.length,
		})
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	for {
		select {
		case <-interrupt:
			log.Println("interrupt")
			close(done)
			wg.Wait()
			return
		}
	}
}
