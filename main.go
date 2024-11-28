package main

import (
	"alexchomiak/go-docker-api/env"
	"alexchomiak/go-docker-api/handlers"
	"alexchomiak/go-docker-api/handlers/stacks"
	"alexchomiak/go-docker-api/handlers/ws"
	"fmt"
	"log"
	"os"

	"github.com/gofiber/contrib/websocket"

	"github.com/gofiber/fiber/v2"
)

func init() {

	sqllite_dir := os.Getenv("SQLLITE_PATH")
	if sqllite_dir == "" {
		sqllite_dir = "./server.db"
	}

	if _, err := os.Stat(env.StacksDir); os.IsNotExist(err) {
		err := os.Mkdir(env.StacksDir, 0755)
		if err != nil {
			panic(fmt.Sprintf("Failed to create stacks directory: %v", err))
		}
	}
}

func main() {
	// Initialize Fiber app
	app := fiber.New()

	// List all running compose clusters (from Docker API)
	app.Get("/api/list", handlers.ListRunningComposeStacks)

	// * Stack APIs
	app.Put("/api/stacks", stacks.ProvisionComposeStack)
	app.Post("/api/stacks", stacks.StartComposeStack)
	app.Delete("/api/stacks", stacks.StopComposeStack)
	app.Get("/api/stacks", stacks.GetStacks)

	// * WebSocket Handlers
	// Middleware to upgrade to WebSocket
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/logs/:id", websocket.New(ws.LogStreamHandler))

	// Determine port (default: 3000)
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Start server
	log.Printf("Server running on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}

}
