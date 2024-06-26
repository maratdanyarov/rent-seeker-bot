package bot

import (
	"bufio"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"rent_seekerbot/internal/real_estate_api"
	"strconv"
	"strings"
	"sync"
)

var (
	bot           *tgbotapi.BotAPI
	userDataMap   = make(map[int64]*UserData)
	userDataMutex sync.Mutex
	zooplaClient  *real_estate_api.ZooplaClient
)

type UserData struct {
	State      string
	ProperType string
	PriceRange string
	Bedrooms   string
	Furnished  string
	Area       string
}

const (
	stateAwaitingPriceRange   = "awaiting_price_range"
	stateAwaitingBedrooms     = "awaiting_bedrooms"
	stateFurnishedUnfurnished = "furnished_unfurnished"
	stateSelectingArea        = "selecting_area"
)

// StartBot initializes and starts the Telegram bot.
func StartBot(token string, zClient *real_estate_api.ZooplaClient) error {
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	zooplaClient = zClient
	if zooplaClient == nil {
		return fmt.Errorf("zooplaClient is nil")
	}

	// Set this to true to log all interactions with telegram servers
	bot.Debug = false

	log.Printf("Authorised on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Create a new cancellable background context. Calling `cancel()` leads to the cancellation of the context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// `updates` is a golang channel which receives telegram updates
	updates := bot.GetUpdatesChan(u)

	// Pass cancellable context to goroutine
	go receiveUpdates(ctx, updates)

	// Tell the user the bot is online
	log.Println("Start listening for updates. Press enter to stop")

	// Wait for a newline symbol, then cancel handling updates
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()

	return nil
}

// receiveUpdates listens for incoming updates and handles them.
func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	// for {` means the loop is infinite until we manually stop it
	for {
		select {
		// stop looping if ctx is cancelled
		case <-ctx.Done():
			return
		// receive update from channel and then handle it
		case update := <-updates:
			handleUpdate(update)
		}
	}
}

// handleUpdate processes incoming updates based on their type.
func handleUpdate(update tgbotapi.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message)
		break
	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
		break
	}
}

func getUserData(chatID int64) *UserData {
	userDataMutex.Lock()
	defer userDataMutex.Unlock()

	userData, exists := userDataMap[chatID]
	if !exists {
		userData = &UserData{}
		userDataMap[chatID] = userData
	}
	return userData
}

// handleMessage processes incoming messages.
func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error

	userData := getUserData(message.Chat.ID)

	if strings.HasPrefix(text, "/") {
		handleCommand(message.Chat.ID, text)
		return
	}
	log.Printf("User state: %s", userData.State)
	switch userData.State {
	case stateAwaitingPriceRange:
		userData.PriceRange = text
		userData.State = stateAwaitingBedrooms

		// Ask for the number of bedrooms
		sendMessageWithMarkup(message.Chat.ID, selectBedroomsMessage, selectBedrooms)
	case stateAwaitingBedrooms:
		userData.Bedrooms = text
		userData.State = stateSelectingArea
		sendMessage(message.Chat.ID, selectArea)
	case stateSelectingArea:
		userData.Area = text
		searchProperties(message.Chat.ID, userData)
	default:
		sendMessage(message.Chat.ID, "I’m sorry, but I don’t recognize this command. Please type /help to see the available list of commands.")
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// handleCommand processes bot commands.
func handleCommand(chatId int64, command string) error {
	var err error

	switch command {
	case "/start":
		userData := getUserData(chatId)
		userData.State = ""
		msg := tgbotapi.NewMessage(chatId, welcomeMessage)
		msg.ReplyMarkup = goButton
		_, err = bot.Send(msg)
	case "/help":
		msg := tgbotapi.NewMessage(chatId, "Hello! I’m here to assist you in finding your perfect home.")
		_, err = bot.Send(msg)
	// ADD MENU OPTION LATER
	default:
		sendMessage(chatId, "I’m sorry, but I don’t recognize this command. Please type /help to see the available list of commands.")
	}

	return err
}

// handleButton proceses callback queries from inline buttons.
func handleButton(query *tgbotapi.CallbackQuery) {
	userData := getUserData(query.Message.Chat.ID)
	switch query.Data {
	case goButtonText:
		sendMessageWithMarkup(query.Message.Chat.ID, selectPropertyMessage, selectProperty)
	case flatButtonText, houseButtonText:
		userData.ProperType = query.Data
		userData.State = stateAwaitingPriceRange
		sendMessage(query.Message.Chat.ID, priceRangeMessage)
	case studioButtonText, oneBedButtonText, twoBedButtonText, threeBedButtonText, fourBedButtonText, fiveBedButtonText:
		userData.Bedrooms = query.Data
		userData.State = stateFurnishedUnfurnished
		sendMessageWithMarkup(query.Message.Chat.ID, selectIsFurnished, isFurnished)
	case furnished, unfurnished:
		userData.Furnished = query.Data
		userData.State = stateSelectingArea
		sendMessage(query.Message.Chat.ID, selectArea)
	}

	bot.Send(tgbotapi.NewCallback(query.ID, ""))
}

func searchProperties(chatID int64, userData *UserData) {
	if zooplaClient == nil {
		log.Println("Error: zooplaClient is nil")
		sendMessage(chatID, "Sorry, I encountered an error while searching for properties. Please try again later.")
		return
	}

	minPrice, maxPrice, err := parsePriceRange(userData.PriceRange)
	if err != nil {
		sendMessage(chatID, "I'm sorry, I couldn't understand the price range. Please try again.")
		return
	}
	bedrooms, err := strconv.Atoi(userData.Bedrooms)
	if err != nil {
		sendMessage(chatID, "I'm sorry, I couldn't understand the number of bedrooms. Please try again.")
		return
	}
	properties, err := zooplaClient.SearchProperties(userData.Area, minPrice, maxPrice, bedrooms, userData.ProperType)
	if err != nil {
		log.Printf("Error searching properties: %v", err)
		sendMessage(chatID, "Sorry, I encountered an error while searching for properties. Please try again later.")
		return
	}
	if len(properties) == 0 {
		sendMessage(chatID, "I'm sorry, but I couldn't find any properties matching your criteria. Please try broadening your search.")
		return
	}

	sendMessage(chatID, fmt.Sprintf("Great! I found %d properties matching your criteria. Here are the top results:", len(properties)))

	for i, property := range properties {
		if i >= 5 {
			break
		}
		propertyMsg := fmt.Sprintf("🏠 %s\n 💰 £%d\n 🛏 %d bedrooms", property.Address, property.Price, property.Bedrooms)
		sendMessage(chatID, propertyMsg)
	}
	sendMessage(chatID, "To start a new search, just type /start")
	userData.State = "" // Reset state after completing the search
}

func parsePriceRange(priceRange string) (int, int, error) {
	parts := strings.Split(priceRange, "-")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid price range format")
	}

	minPrice, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}

	maxPrice, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}

	return minPrice, maxPrice, nil
}

// sendMessageWithMarkup edits a message with new text and markup/
func sendMessageWithMarkup(chatID int64, text string, markup tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = markup
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message with new text and markup: %v", err)
	}
}

// sendMessage sends a message with new text
func sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("Failed to send message with new text and markup: %v", err)
	}
}
