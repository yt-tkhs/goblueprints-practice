package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"github.com/labstack/gommon/log"
)

type room struct {
	// channel that holds messages for transfer to other clients.
	forward chan []byte

	// client trying to join
	join    chan *client

	// client trying to leave
	leave   chan *client

	// clients that joined the room
	clients map[*client]bool

	// history of message
	messages [][]byte
}

func newRoom() *room {
	return &room {
		forward: make(chan []byte),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
		messages:make([][]byte, 0, 10000),
	}
}

const (
	socketBufferSize = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:     socketBufferSize,
	WriteBufferSize:    messageBufferSize,
}

func (r *room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP: ", err)
		return
	}

	client := &client {
		socket: socket,
		send: make(chan []byte, messageBufferSize),
		room: r,
	}

	// send past messages
	for _, msg := range r.messages {
		client.send <- msg
	}

	r.join <- client

	defer func() {
		r.leave <- client
	}()

	go client.write()
	client.read()
}

func (r *room) run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
		case msg := <-r.forward:
			r.addHistory(msg)
			for client := range r.clients {
				select {
				case client.send <- msg:
					// メッセージを送信
				default:
					// 送信に失敗したクライアントは退室させる
					delete(r.clients, client)
					close(client.send)
				}
			}
		}
	}
}

func (r *room) addHistory(msg []byte)  {
	r.messages = append(r.messages, msg)
}