package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("application running")
	InitRouter()

}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  8000000,
	WriteBufferSize: 8000000,
	CheckOrigin:     func(r *http.Request) bool { return true }, // TBD: actually check origin
}
func InitRouter() *mux.Router {

	log.Println("router started...")

	addr := ":8081"

	router := mux.NewRouter()
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// create the header for the response
		w.Header().Set("Content-Type","application/json;charset=utf8")

		fmt.Fprint(w, "Hello, Welcome to Websocket Test api")
	})

	router.HandleFunc("/ws", ServeWs)




	methodsOk := handlers.AllowedMethods([]string{
		"GET",
		"PUT",
		"POST",
		"OPTIONS",
	})
	headersOk := handlers.AllowedHeaders([]string{
		"Authorization",
		"X-Csrf-Token",
		"X-Xsrf-Token",
	})
	originsOk := handlers.AllowedOrigins([]string{
		"http://localhost",
		"http://127.0.0.1",
		"http://localhost:4200",
	})
	allowCredentials := handlers.AllowCredentials()

	err := http.ListenAndServe(addr, handlers.CORS(originsOk, headersOk, methodsOk, allowCredentials)(router))
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
	return router
}
const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 8000000
)

func ServeWs(w http.ResponseWriter, r *http.Request) {
	var conn *websocket.Conn

	if _, ok := w.(http.Hijacker); !ok {

	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("upgrage failed: %s", err)
	}
	messageChan := make(chan string)
	go ReadPump(conn, messageChan)
	go WritePump(conn, messageChan)
}
func ReadPump(conn *websocket.Conn, messageChan chan string) {

	defer func() {

		log.Print("gonna close the connection")
		conn.Close()
	}()
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error { conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {

		//log.Print(user.Conn)
		_, msg, err := conn.ReadMessage()
		log.Println("message", msg)
		if err != nil {

			log.Printf("an error %s", err)
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		messageChan <- "feedback: server received message"
		log.Println("Message read")

	}
}

func WritePump(conn *websocket.Conn, messageChan chan string) {

	defer func() {

		conn.Close()

	}()
	for {
		select {
		case message, ok := <-messageChan:
			log.Println("value passed into channel")
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			d, err := json.Marshal(message)
			if err != nil {
				return
			}
			w.Write(d)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}
