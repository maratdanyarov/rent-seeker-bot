package main

import (
	"log"
	"rent_seekerbot/internal/bot"
	"rent_seekerbot/internal/config"
	"rent_seekerbot/internal/real_estate_api"
)

func main() {
	config.LoadConfig()

	token := config.GetEnv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}
	zooplaClientID := config.GetEnv("ZOOPLA_CLIENT_ID")
	if zooplaClientID == "" {
		log.Fatal("ZOOPLA_CLIENT_ID must be set")
	}
	zooplaClientSecret := config.GetEnv("ZOOPLA_CLIENT_SECRET")
	if token == "" || zooplaClientID == "" || zooplaClientSecret == "" {
		log.Fatal("ZOOPLA_CLIENT_SECRET must be set")
	}

	// Create Zoopla client
	zooplaClient := real_estate_api.NewZooplaClient(zooplaClientID, zooplaClientSecret, "RentSeekerBot")

	// Test Zoopla API connection
	log.Println("Testing Zoopla API connection...")
	err := zooplaClient.TestApiConnection()
	if err != nil {
		log.Fatalf("Zoopla API test failed: %v", err)
	}
	log.Printf("Zoopla API test successful!")

	// Start the bot
	err = bot.StartBot(token, zooplaClient)
	if err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
