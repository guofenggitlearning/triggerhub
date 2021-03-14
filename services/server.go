package services

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"

	fiber "github.com/gofiber/fiber/v2"
	websocket "github.com/gofiber/websocket/v2"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

type ActionStatus int

const (
	ActionIdle ActionStatus = iota
	ActionRunning
	ActionEnded
	ActionFailed
)

type ResultMessage struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}
type RelayedAction struct {
	Action string `json:"action"`
	Token  string `json:"token"`
}
type ListenerMessage struct {
	Command string   `json:"command"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}
type Listener struct {
	name       string
	connection *websocket.Conn
	msgType    int
	actions    []struct {
		name   string
		status ActionStatus
	}
	ip string
}

var listeners []Listener

// Serve starts a server and listens for requests
func Serve() {
	var port = viper.GetInt("port")
	var useTLS = viper.GetBool("tls")
	var cert = viper.GetString("cert")
	var key = viper.GetString("key")
	var token = viper.GetString("token")

	if token == "" {
		log.Fatalln("The dispatcher token cannot be empty")
	}

	// info
	if useTLS {
		if cert == "" || key == "" {
			log.Fatal("The certificate and key file are needed to run with TLS enabled")
		}
		log.Println("TLS enabled")
	}

	// Service set up
	app := fiber.New()

	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Set("Access-Control-Allow-Origin", "*")
		ctx.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		return ctx.Next()
	})

	app.Options("*", func(ctx *fiber.Ctx) error {
		ctx.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		return ctx.SendStatus(fiber.StatusOK)
	})

	// listeners pushing actions
	app.Get("/actions/:action/:token", handleGet)

	// Services listening to us
	app.Use("/ws", handleRequireWsUpgrade)
	app.Get("/ws/:token", websocket.New(handleWsClient))

	// Main listener
	if useTLS {
		// Read TLS certificate
		cer, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatal(err)
		}

		tlsConfig := &tls.Config{Certificates: []tls.Certificate{cer}}
		addr := fmt.Sprintf(":%d", port)

		// Create custom listener
		ln, err := tls.Listen("tcp", addr, tlsConfig)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Listening TLS on", addr)
		log.Fatal(app.Listener(ln))
	} else {
		addr := fmt.Sprintf(":%d", port)
		log.Println("Listening HTTP on", addr)
		log.Fatal(app.Listen(addr))
	}
}

// handleGet handles the request to run a certain trigger
func handleGet(ctx *fiber.Ctx) error {
	var action = ctx.Params("action")
	var token = ctx.Params("token")

	if action == "" || token == "" {
		ctx.Status(fiber.StatusNotFound)
		return ctx.SendString("Not found")
	}

	found := false
	for i := 0; i < len(listeners); i++ {
		for j := 0; j < len(listeners[i].actions); j++ {
			if listeners[i].actions[j].name == action {
				found = true
				notifyListener(listeners[i], action, token, listeners[i].name)

				// Many listeners could declare the same action name, navigate all
				continue
			}
		}
	}
	if !found {
		log.Println("Not found", ctx.Path())
		ctx.Status(fiber.StatusNotFound)
		return ctx.SendString("Not found")
	}

	return nil
}

func handleRequireWsUpgrade(c *fiber.Ctx) error {
	// IsWebSocketUpgrade returns true if the client
	// requested upgrade to the WebSocket protocol.
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		c.Locals("IP", c.IP())
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func handleWsClient(c *websocket.Conn) {
	// c.Locals is added to the *websocket.Conn
	if c.Locals("allowed") != true {
		log.Println("[WS] Closing non-upgraded connection")
		c.WriteJSON(ResultMessage{Ok: false, Message: "Please, upgrade to web sockets"})
		c.Close()
		return
	} else if c.Params("token") != viper.GetString("token") {
		log.Println("[WS] Unauthorized token", c.Params("token"))
		c.WriteJSON(ResultMessage{Ok: false, Message: "Unauthorized"})
		c.Close()
		return
	}

	log.Println("[WS] Connection from", c.Locals("IP"))

	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
	var (
		msgType    int
		rawMessage []byte
		err        error
	)
	for {
		// Read raw
		if msgType, rawMessage, err = c.ReadMessage(); err != nil {
			log.Println("[WS]", err)
			break
		}

		handleIncomingMessage(c, rawMessage, msgType)

	}

	log.Println("[WS] Disconnected", c.Locals("IP"))
	removeListener(c)
}

// Helpers

func handleIncomingMessage(c *websocket.Conn, rawMessage []byte, msgType int) {
	var decodedMessage ListenerMessage

	if err := json.Unmarshal(rawMessage, &decodedMessage); err != nil {
		log.Println("[WS] Could not unmarshall the message", err)
		log.Println("[WS] Disconnecting", c.Locals("IP"))
		c.WriteJSON(ResultMessage{Ok: false})
		c.Close()
		return
	}

	// Special listener commands
	switch decodedMessage.Command {
	case "register":
		err := addListener(decodedMessage, c, msgType, c.Locals("IP").(string))
		if err != nil {
			c.WriteJSON(ResultMessage{Ok: false, Message: err.Error()})
		} else {
			c.WriteJSON(ResultMessage{Ok: true})
		}
		break
	default:
		if decodedMessage.Command != "" {
			log.Printf("[WS] Unrecognized command from %s: %s", c.Locals("IP"), decodedMessage.Command)
			c.WriteMessage(msgType, []byte(`{"ok":false,"message":"Unrecognized command"}`))
		}
	}
}

func addListener(decodedMessage ListenerMessage, connection *websocket.Conn, msgType int, ip string) error {
	// Check for duplicates
	for i := 0; i < len(listeners); i++ {
		if listeners[i].name == decodedMessage.Name && listeners[i].ip == ip {
			return errors.Errorf("A service with the same name is already registered from this host")
		}
	}

	actions := make([]struct {
		name   string
		status ActionStatus
	}, len(decodedMessage.Actions))

	for i := 0; i < len(decodedMessage.Actions); i++ {
		actions[i].name = decodedMessage.Actions[i]
		actions[i].status = ActionIdle
	}

	newListener := Listener{
		name:       decodedMessage.Name,
		connection: connection,
		msgType:    msgType,
		actions:    actions,
		ip:         ip,
	}
	listeners = append(listeners, newListener)

	log.Printf("[WS] Registered %s (%s)", decodedMessage.Name, ip)
	return nil
}

func removeListener(c *websocket.Conn) {
	name := ""
	ip := ""

	// Remove from the list
	idx := -1
	for i := 0; i < len(listeners); i++ {
		if listeners[i].connection == c {
			idx = i
			name = listeners[i].name
			ip = listeners[i].ip
			c.Close()
			break
		}
	}

	if idx < 0 {
		// log.Println("[err] Listener item not found")
		return
	}

	if len(listeners) > 0 {
		listeners[len(listeners)-1], listeners[idx] = listeners[idx], listeners[len(listeners)-1]
	}
	listeners = listeners[:len(listeners)-1]
	log.Printf("[WS] Unregistered %s (%s)", name, ip)
}

func notifyListener(listener Listener, action string, token string, listenerName string) {
	log.Println("[http] Relaying", action, "to", listenerName)
	err := listener.connection.WriteJSON(RelayedAction{Action: action, Token: token})
	if err != nil {
		log.Printf("[%s] Error: %v", listener.name, err)
		return
	}

	var response ResultMessage
	err = listener.connection.ReadJSON(&response)
	if err != nil {
		log.Printf("[%s] Error: %v", listener.name, err)
	}
	log.Println("[result]", response)
}
