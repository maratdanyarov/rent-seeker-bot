package main

import (
	"log"
	"rent_seekerbot/internal/bot"
	"rent_seekerbot/internal/config"
	"rent_seekerbot/internal/database"
	"rent_seekerbot/internal/real_estate_api"
)

func main() {
	config.LoadConfig()

	token := config.GetEnv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN must be set")
	}
	db, err := database.NewDB("rent_seeker.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.CreateTables()
	if err != nil {
		log.Fatal(err)
	}

	zooplaClient := real_estate_api.NewZooplaClient(
		config.GetEnv("ZOOPLA_CLIENT_ID"),
		config.GetEnv("ZOOPLA_CLIENT_SECRET"),
		config.GetEnv("ZOOPLA_AGENCY_REF"),
	)

	// Test Zoopla API connection
	log.Println("Testing Zoopla API connection...")
	err = zooplaClient.TestApiConnection()
	if err != nil {
		log.Fatalf("Zoopla API test failed: %v", err)
	}
	log.Printf("Zoopla API test successful!")

	// Start the bot
	err = bot.StartBot(token, zooplaClient, db)
	if err != nil {
		log.Fatalf("Failed to start bot: %v", err)
	}
}
