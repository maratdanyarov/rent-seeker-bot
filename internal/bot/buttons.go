package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

const (
	// Button text
	goButtonText = "Let's go!"

	flatButtonText  = "Flat"
	houseButtonText = "House"

	studioButtonText   = "Studio"
	oneBedButtonText   = "1"
	twoBedButtonText   = "2"
	threeBedButtonText = "3"
	fourBedButtonText  = "4"
	fiveBedButtonText  = "5"

	furnished   = "Furnished"
	unfurnished = "Unfurnished"
)

// Create the "Let's go" button
var goButton = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData(goButtonText, goButtonText)))

// Create "Select the property type" button
var selectProperty = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData(flatButtonText, flatButtonText)),
	tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(houseButtonText, houseButtonText)))

// Create "Select the number of bedrooms" buttons
var selectBedrooms = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData(studioButtonText, studioButtonText),
	tgbotapi.NewInlineKeyboardButtonData(oneBedButtonText, oneBedButtonText),
	tgbotapi.NewInlineKeyboardButtonData(twoBedButtonText, twoBedButtonText)),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData(threeBedButtonText, threeBedButtonText),
		tgbotapi.NewInlineKeyboardButtonData(fourBedButtonText, fourBedButtonText),
		tgbotapi.NewInlineKeyboardButtonData(fiveBedButtonText, fiveBedButtonText)),
)

// Create "Select furnished or unfurnished" buttons
var isFurnished = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(
	tgbotapi.NewInlineKeyboardButtonData(furnished, furnished),
	tgbotapi.NewInlineKeyboardButtonData(unfurnished, unfurnished)))
