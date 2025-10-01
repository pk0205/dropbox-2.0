package main

import (
	"context"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/pk0205/dropbox-2.0/db"
)

func main() {
		app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowCredentials: true,
	}))

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	PORT := os.Getenv("PORT")

	conn, err := db.Connect()
    if err != nil {
        log.Fatal("Error connecting DB:", err)
    }
    defer conn.Close(context.Background())

    if err := db.PingDB(conn); err != nil {
        log.Fatal("Unable to ping database:", err)
    }

    if err := db.SetupDB(conn); err != nil {
        log.Fatal("Unable to setup database:", err)
    }


	app.Get("/", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{"msg": "Hello, World!!!"})
	})

	log.Fatal(app.Listen(":" + PORT))

}