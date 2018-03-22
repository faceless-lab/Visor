package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"time"
)

const port string = ":8787" // port number
const bufSize int = 4096    // websocket buffer size
const fps byte = 60
const timeout = 50 * time.Millisecond

var frameQueue = make(chan []byte, fps)

var upgrader = websocket.Upgrader{
	ReadBufferSize:    bufSize,
	WriteBufferSize:   bufSize,
	EnableCompression: true,
}

func index(rw http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/index.html")
	t.Execute(rw, nil)
}

func ws(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, nil)
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	}
	defer conn.Close()

	for {
		_, buf, err := conn.ReadMessage()
		if err != nil {
			rw.Write([]byte(err.Error()))
			return
		}

		select {
		case frameQueue <- buf:
		case <-time.After(timeout):
			continue
		}

	}

}

func screen(rw http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(rw, r, rw.Header())
	if err != nil {
		rw.Write([]byte(err.Error()))
		return
	}

	go func() {
		defer conn.Close()

		for {
			select {
			case buf := <-frameQueue:
				if err := conn.WriteMessage(websocket.BinaryMessage, buf); err != nil {
					continue
				}
			case <-time.After(timeout):
				continue
			}

		}
	}()
}

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/ws", ws)
	r.HandleFunc("/screen", screen)
	r.HandleFunc("/", index)

	r.PathPrefix("/").
		Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))

	log.Printf("Starting server at %s", port)
	if err := http.ListenAndServe(port, r); err != nil {
		log.Fatal(err)
	}
}
