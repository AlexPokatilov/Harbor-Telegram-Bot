package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Resource represents the resource associated with the event
type Resource struct {
	Digest      string `json:"digest"`
	Tag         string `json:"tag"`
	ResourceURL string `json:"resource_url"`
}

// Repository displays information about the repository
type Repository struct {
	DateCreated  int64  `json:"date_created"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}

// EventData contains event details
type EventData struct {
	Resources  []Resource `json:"resources"`
	Repository Repository `json:"repository"`
}

// WebhookPayload represents the body of the webhook request
type WebhookPayload struct {
	Type      string    `json:"type"`
	OccurAt   int64     `json:"occur_at"`
	Operator  string    `json:"operator"`
	EventData EventData `json:"event_data"`
}

type SendMessageParams struct {
	ChatID  int64
	Message string
	TopicID *int64
}

var bot *tgbotapi.BotAPI

func initTelegramBot() {
	var err error
	bot, err = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	debugModeStr := os.Getenv("DEBUG_MODE") //true - hide data, false - show data
	debugModeBool, err := strconv.ParseBool(debugModeStr)
	if err != nil {
		log.Printf("ERROR!!! When converting DEBUG_MODE to bool: %v", err)
		return
	}

	bot.Debug = debugModeBool
	log.Printf("OK! Connected to telegram bot account: https://t.me/%s", bot.Self.UserName)
}

func extractDomain(resourceURL string) string {
	start := strings.Index(resourceURL, "//") // Find the index where the double slashes begin "//"
	if start == -1 {
		return "" // URL does not contain "//"
	}

	start += 2                                     // Index shift beyond limits "//"
	end := strings.Index(resourceURL[start:], "/") // Find the index of the first slash "/" after "//"
	if end == -1 {
		return resourceURL[start:] // URL does not contain an extra "/"
	}

	return resourceURL[start : start+end] // Return the substring between start and end
}

func formatMessage(payload WebhookPayload) string {
	// Checking if resources are available in the payload
	if len(payload.EventData.Resources) == 0 {
		return "WARNING!! Webhook Event Received, but no resources are available."
	}

	resource := payload.EventData.Resources[0] // We use the first resource
	repo := payload.EventData.Repository

	harborURL := strings.Split(resource.ResourceURL, "/")[0]
	harborLink := fmt.Sprintf("https://%s/harbor/projects", harborURL)
	harborChartURL := extractDomain(resource.ResourceURL)
	harborChartLink := fmt.Sprintf("https://%s/harbor/projects", harborChartURL)

	var message string
	switch payload.Type {
	case "PUSH_ARTIFACT":
		message = fmt.Sprintf("\nNew 🐳 image pushed by: <b>%s</b>\n", payload.Operator)
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", harborLink, harborURL)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", repo.Namespace)
		message += fmt.Sprintf("• Repository: <b>%s</b>\n", repo.RepoFullName)
		message += fmt.Sprintf("• Tag: <b>%s</b>", resource.Tag)
	case "UPLOAD_CHART":
		message = fmt.Sprintf("\nNew ☸️ chart version uploaded by: <b>%s</b>\n", payload.Operator)
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", harborChartLink, harborChartURL)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", repo.Namespace)
		message += fmt.Sprintf("• Chart: <b>%s</b>\n", repo.Name)
		message += fmt.Sprintf("• Version: <b>%s</b>", resource.Tag)
	default:
		message = "WARNING!! Received an unknown event type."
	}

	return message
}

func toJSONPretty(v interface{}) string {
	prettyJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error when marshalling to pretty JSON: %v", err)
		return ""
	}
	return string(prettyJSON)
}

func sendTelegramMessage(params SendMessageParams) {
	msg := tgbotapi.NewMessage(params.ChatID, params.Message)

	if params.TopicID != nil {
		msg.ReplyToMessageID = int(*params.TopicID)
	}

	msg.ParseMode = "HTML"
	response, err := bot.Send(msg) // Sending a message
	if err != nil {
		log.Printf("Error when sending message: %v", err)
	} else {
		log.Printf("Endpoint: sendMessage, params: map[chat_id:%d parse_mode:HTML text:\n%s]\n", params.ChatID, params.Message)
		// Using toJSONPretty to format the response in pretty JSON
		prettyJSON := toJSONPretty(response)
		log.Printf("Endpoint: sendMessage, response:\n%s\n", prettyJSON)
	}
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
	// Check for a POST request
	if r.Method != "POST" {
		http.Error(w, "ERROR!!! Only POST method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	// Decode the JSON from the request body
	var payload WebhookPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	chatIdStr := os.Getenv("CHAT_ID")
	chatIdInt, err := strconv.ParseInt(chatIdStr, 10, 64)
	if err != nil {
		log.Printf("ERROR!!! When converting chat ID to int64: %v", err)
		return
	}

	topicIDStr := os.Getenv("TOPIC_ID")
	var topicIDPtr *int64 // 使用指针类型以支持可选性

	if topicIDStr != "" {
		topicID, err := strconv.ParseInt(topicIDStr, 10, 64)
		if err != nil {
			log.Printf("ERROR!!! When converting topic ID to int64: %v\", err", err)
			return
		}
		topicIDPtr = &topicID
	}

	// Forming and sending a message in Telegram
	message := formatMessage(payload) // Using formatMessage to create a message
	sendTelegramMessage(SendMessageParams{
		ChatID:  chatIdInt,
		Message: message,
		TopicID: topicIDPtr,
	})

	// Response to request
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK! Webhook received."))
}

func main() {
	initTelegramBot() // Initialize the Telegram bot
	http.HandleFunc("/webhook-bot", handleWebhook)
	log.Fatal(http.ListenAndServe(":441", nil))
}
