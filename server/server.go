package server

import (
	"crypto/tls"
	"fmt"
	"log"

	fiber "github.com/gofiber/fiber/v2"
)

// Run starts a server and listens for requests
func Run(port int, authToken string, useTLS bool, cert string, key string) {
	// info
	log.Output(1, fmt.Sprintf("Running with token %s", authToken))

	if useTLS {
		if cert == "" || key == "" {
			log.Fatal("The certificate and key file are needed to run with TLS enabled")
		}
		log.Output(1, "TLS enabled")
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

	app.Get("/backup/:id/:token", handleGet)

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

		log.Output(1, fmt.Sprintf("Listening TLS on %s\n", addr))
		log.Fatal(app.Listener(ln))
	} else {
		addr := fmt.Sprintf(":%d", port)
		log.Output(1, fmt.Sprintf("Listening HTTP on %s\n", addr))
		log.Fatal(app.Listen(addr))
	}
}

// handleGet handles the request to run a certain trigger
func handleGet(ctx *fiber.Ctx) error {
	log.Output(1, ctx.Params("id"))
	log.Output(1, ctx.Params("token"))
	return nil
}
