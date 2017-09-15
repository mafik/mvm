package mvm

import (
	"fmt"
	"golang.org/x/net/websocket"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func Start() {
	fmt.Println("Loading VM image...")
	err := LoadImage()
	if err != nil {
		fmt.Println("Error while loading VM image:", err)
		os.Exit(1)
	}
	fmt.Println("VM image loaded successfully")
	fmt.Println("Starting the VM and WebGUI")

	events := make(chan Event)
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		<-signals
		var e Event
		e.Type = "Interrupt"
		events <- e
	}()
	updates := MakeFanOut()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Requested %s\n", r.RequestURI)
		switch r.RequestURI {
		case "/":
			index, err := ioutil.ReadFile("static/index.html")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
			script, err := ioutil.ReadFile("static/script.js")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
			_, err = fmt.Fprintf(w, "%s<script>%s</script>", index, script)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
		case "/favicon.ico":
			file, err := os.Open("static/favicon.ico")
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprint(w, err)
				return
			}
			_, err = io.Copy(w, file)
			if err != nil {
				fmt.Printf("Error: %v", err)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		ws_updates := updates.Open()
		go func() {
			for update := range ws_updates {
				err := websocket.Message.Send(ws, update)
				if err != nil {
					fmt.Println(err)
					updates.Close(ws_updates)
				}
			}
		}()
		for {
			var e Event
			err := websocket.JSON.Receive(ws, &e)
			if err != nil {
				fmt.Println(err)
				updates.Close(ws_updates)
				return
			}
			events <- e
		}
	}))

	go func() {
		log.Fatal(http.ListenAndServe(":8000", nil))
	}()

	for keep_running {
		select {
		case event := <-events:
			ProcessEvent(event, updates.in)
			continue
		default:
		}
		select {
		case event := <-events:
			ProcessEvent(event, updates.in)
		case task := <-tasks:
			task.Run(events)
		}
	}
}
