package mvm

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"

	"golang.org/x/net/websocket"
)

type HttpClient struct {
	events   chan Event
	sync_in  chan Event
	sync_out chan string
}

func (c *HttpClient) Call(request string) (response Event, err error) {
	c.sync_out <- request
	var ok bool
	response, ok = <-c.sync_in
	if !ok {
		err = errors.New("Couldn't send the request")
	}
	return
}

type EventClient struct {
	event  Event
	client Client
}

func Start() {
	fmt.Println("Loading VM image...")
	err := LoadImage()
	if err != nil {
		fmt.Println("Error while loading VM image:", err)
		os.Exit(1)
	}
	fmt.Println("VM image loaded successfully")
	fmt.Println("Starting the VM and WebGUI")

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	quit := make(chan struct{})
	main_chan := make(chan EventClient)
	orphan_events := make(chan Event)
	go func() {
		for {
			select {
			case e := <-orphan_events:
				main_chan <- EventClient{e, nil}
			case <-quit:
				break
			}
		}
	}()

	go func() {
		select {
		case <-signals:
			var e Event
			e.Type = "Interrupt"
			orphan_events <- e
		case <-quit:
		}
	}()

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

	new_events := make(chan chan Event)
	new_sync_in := make(chan chan Event)
	new_sync_out := make(chan chan string)

	http.Handle("/events", websocket.Handler(func(ws *websocket.Conn) {
		events := make(chan Event)
		new_events <- events
		for {
			var e Event
			err := websocket.JSON.Receive(ws, &e)
			if err != nil {
				fmt.Println("Events closed: ", err)
				close(events)
				return
			}
			select {
			case <-quit:
				break
			default:
			}
			select {
			case events <- e:
				continue
			case <-quit:
				break
			}
		}
	}))

	http.Handle("/sync", websocket.Handler(func(ws *websocket.Conn) {
		sync_in := make(chan Event)
		sync_out := make(chan string)
		new_sync_in <- sync_in
		new_sync_out <- sync_out
		go func() {
			for {
				select {
				case req := <-sync_out:
					err := websocket.Message.Send(ws, req)
					if err != nil {
						fmt.Println("Sync (out) closed:", err)
						close(sync_out)
					}
				case <-quit:
					break
				}
			}
		}()
		for {
			var e Event
			err := websocket.JSON.Receive(ws, &e)
			if err != nil {
				fmt.Println("Sync (in) closed:", err)
				close(sync_in)
				return
			}
			select {
			case sync_in <- e:
			case <-quit:
				break
			}
		}
	}))

	go func() {
		log.Fatal(http.ListenAndServe("localhost:8000", nil))
	}()

	// Main loop
	go func() {
		for {
			select { // give priority to main_chan
			case ec := <-main_chan:
				ProcessEvent(ec.event, ec.client)
			default:
				select {
				case ec := <-main_chan:
					ProcessEvent(ec.event, ec.client)
				case task := <-tasks:
					task.Run(orphan_events)
				}
			}
			if !keep_running {
				close(quit)
				return
			}
		}
	}()

	for {
		var events chan Event
		select {
		case events = <-new_events:
		case <-quit:
			return
		}

		var sync_in chan Event
		select {
		case sync_in = <-new_sync_in:
		case <-quit:
			return
		}

		var sync_out chan string
		select {
		case sync_out = <-new_sync_out:
		case <-quit:
			return
		}

		fmt.Println("New client connected")

		go func() {
			client := HttpClient{events, sync_in, sync_out}
			for {
				var e Event
				var ok bool
				select {
				case e, ok = <-events:
					if !ok {
						fmt.Println("Client disconnected")
						return
					}
				case <-quit:
					return
				}

				select {
				case main_chan <- EventClient{e, &client}:
				case <-quit:
					return
				}
			}
		}()
	}
}
