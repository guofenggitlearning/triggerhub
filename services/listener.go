package services

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/brickpop/triggerhub/config"

	fiber "github.com/gofiber/fiber/v2"
	websocket "github.com/gofiber/websocket/v2"
	"github.com/spf13/viper"
)

var actions []config.ActionEntry

// Listen starts a server and listens for requests
func Listen() {
	var configFile = viper.GetString("config")
	var port = viper.GetInt("port")
	var useTLS = viper.GetBool("tls")
	var cert = viper.GetString("cert")
	var key = viper.GetString("key")

	// info
	if useTLS {
		if cert == "" || key == "" {
			log.Fatal("The certificate and key file are needed to run with TLS enabled")
		}
		log.Println("TLS enabled")
	}

	// info
	if configFile != "" {
		log.Println("Using config file", configFile)
	}

	// config entries
	err := viper.UnmarshalKey("actions", &actions)
	if err != nil {
		log.Fatal(err)
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

	// Clients pushing triggers
	app.Get("/triggers/:id/:token", handleGet)

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

// // handleGet handles the request to run a certain trigger
// func handleGet(ctx *fiber.Ctx) error {
// 	var id = ctx.Params("id")
// 	var token = ctx.Params("token")

// 	if id == "" || token == "" {
// 		ctx.Status(fiber.StatusNotFound)
// 		return ctx.SendString("Not found")
// 	}

// 	found := false
// 	for i := 0; i < len(paths); i++ {
// 		if paths[i].ID == id {
// 			if paths[i].Token == token {
// 				found = true
// 				break
// 			}
// 			log.Println("Not found", ctx.Path())
// 			ctx.Status(fiber.StatusNotFound)
// 			return ctx.SendString("Not found")
// 		}
// 	}
// 	if !found {
// 		log.Println("Not found", ctx.Path())
// 		ctx.Status(fiber.StatusNotFound)
// 		return ctx.SendString("Not found")
// 	}

// 	log.Println("TO DO: Handle request", id, token)
// 	return nil
// }

// func handleRequireWsUpgrade(c *fiber.Ctx) error {
// 	// IsWebSocketUpgrade returns true if the client
// 	// requested upgrade to the WebSocket protocol.
// 	if websocket.IsWebSocketUpgrade(c) {
// 		c.Locals("allowed", true)
// 		c.Locals("IP", c.IP())
// 		return c.Next()
// 	}
// 	return fiber.ErrUpgradeRequired
// }

// func handleWsClient(c *websocket.Conn) {
// 	// c.Locals is added to the *websocket.Conn
// 	if c.Locals("allowed") != true {
// 		log.Println("Closing non-upgraded connection")
// 		c.WriteMessage(1, []byte(`{"error":true,"message":"Please, upgrade to web sockets"}`))
// 		c.Close()
// 		return
// 	} else if c.Params("token") != "1234" {
// 		log.Println("Unauthorized token", c.Params("token"))
// 		c.WriteMessage(1, []byte(`{"error":true,"message":"Unauthorized"}`))
// 		c.Close()
// 		return
// 	}

// 	log.Println("Connection from", c.Locals("IP"))

// 	// websocket.Conn bindings https://pkg.go.dev/github.com/fasthttp/websocket?tab=doc#pkg-index
// 	var (
// 		msgType int
// 		msg     []byte
// 		err     error
// 	)
// 	for {
// 		if msgType, msg, err = c.ReadMessage(); err != nil {
// 			log.Println("[read]", err)
// 			break
// 		}
// 		// log.Printf("recv: %s", msg)

// 		if err = c.WriteMessage(msgType, msg); err != nil {
// 			log.Println("[write]", err)
// 			break
// 		}
// 	}

// 	log.Println("Disconnected", c.Locals("IP"))
// }
