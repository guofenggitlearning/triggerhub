package server

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/brickpop/packerd/config"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
	"github.com/walle/targz"
)

var paths []config.PathEntry

// Run starts a server and listens for requests
func Run() {
	var configFile = viper.GetString("config")
	var port = viper.GetInt("port")
	var authToken = viper.GetString("token")
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
	log.Println("Running with token", authToken)

	// config entries
	err := viper.UnmarshalKey("paths", &paths)
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
	var id = ctx.Params("id")
	var token = ctx.Params("token")

	if id == "" || token == "" {
		ctx.Status(fiber.StatusNotFound)
		return ctx.SendString("Not found")
	}

	outputFile := fmt.Sprintf("/tmp/packerd-%d.tar.gz", time.Now().UnixNano())

	srcPath := ""
	found := false
	for i := 0; i < len(paths); i++ {
		if paths[i].ID == id {
			if paths[i].Token == token {
				found = true
				srcPath = paths[i].Path
				break
			}
			ctx.Status(fiber.StatusNotFound)
			return ctx.SendString("Not found")
		}
	}
	if !found {
		ctx.Status(fiber.StatusNotFound)
		return ctx.SendString("Not found")
	}

	log.Println(fmt.Sprintf("[%s] Bundling %s into %s", id, srcPath, outputFile))
	err := targz.Compress(srcPath, outputFile)
	if err != nil {
		ctx.Status(fiber.StatusInternalServerError)
		log.Println(err)
		return ctx.SendString("Internal server error")
	}

	err = ctx.SendFile(outputFile, false)
	if err != nil {
		return err
	}

	err = os.Remove(outputFile)
	return err
}
