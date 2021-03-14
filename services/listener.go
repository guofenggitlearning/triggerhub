package services

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/brickpop/triggerhub/config"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var actions []config.ActionEntry
var dispatcher struct {
	host  string
	tls   bool
	token string
}

// Listen connects to the dispatcher and listens for actions
func Listen() {
	readConfig()

	// Signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Connect
	var scheme string
	if dispatcher.tls {
		scheme = "wss"
	} else {
		scheme = "ws"
	}
	u := url.URL{Scheme: scheme, Host: *&dispatcher.host, Path: fmt.Sprintf("/ws/%s", dispatcher.token)}
	log.Printf("[listener] Connecting to %s", dispatcher.host)

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("[listener] Dial:", err)
	}
	defer c.Close()

	sendRegisterMessage(c)

	done := make(chan struct{})

	// Processing loop
	go func() {
		defer close(done)
		for {
			var message RelayedAction
			err := c.ReadJSON(&message)
			if err != nil {
				log.Println("[read]", err)
				return
			}
			// Launch
			err = handleIncomingTrigger(message)
			if err != nil {
				log.Println("[result]", err)
				err = c.WriteJSON(ResultMessage{Ok: false, Message: err.Error()})
				continue
			}
			err = c.WriteJSON(ResultMessage{Ok: true})
			if err != nil {
				log.Println("[result]", err)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		// case t := <-ticker.C:
		// 	err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
		// 	if err != nil {
		// 		log.Println("write:", err)
		// 		return
		// 	}
		case <-interrupt:
			log.Println("[listener] Received SIGINT")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("[listener]:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func sendRegisterMessage(c *websocket.Conn) {
	name := viper.GetString("name")

	err := c.WriteJSON(ListenerMessage{Command: "register", Name: name, Actions: []string{"a", "b", "c"}})
	if err != nil {
		log.Println("[register]:", err)
		return
	}

	var response ResultMessage
	err = c.ReadJSON(&response)
	if err != nil {
		log.Fatalln("[register]", err)
	} else if !response.Ok {
		log.Fatalln("[register]", response.Message)
	}
	log.Println("[register] Connected")
}

func handleIncomingTrigger(action RelayedAction) error {
	log.Println("[request]", action)

	for i := 0; i < len(actions); i++ {
		if actions[i].Name == action.Action {
			if actions[i].Token != action.Token {
				return errors.Errorf("Invalid token")
			}

			return launchActionCommand(actions[i].Command)
		}
	}

	return errors.Errorf("Not found")
}

func launchActionCommand(command string) error {

	log.Println("DO ACTION", command)
	return nil
}

// HELPERS

func readConfig() {
	var configFile = viper.GetString("config")

	// info
	if configFile == "" {
		log.Fatal("[config] The config file is required")
	}

	// config entries
	err := viper.UnmarshalKey("actions", &actions)
	if err != nil {
		log.Fatal(err)
	} else if len(actions) == 0 {
		log.Fatal("[config] No actions are defined")
	}

	dispatcher.host = viper.GetString("dispatcher.host")
	dispatcher.tls = viper.GetBool("dispatcher.tls")
	dispatcher.token = viper.GetString("dispatcher.token")

	if dispatcher.host == "" {
		log.Fatal("[config] No dispatcher host defined")
	} else if dispatcher.token == "" {
		log.Fatal("[config] No dispatcher token defined")
	}
}
