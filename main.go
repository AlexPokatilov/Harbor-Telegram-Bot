package main

import (
    "strconv"
    "os"
    "strings"
    "fmt"
    "encoding/json"
    "log"
    "net/http"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Resource represents the resource associated with the event
type Resource struct {
    Digest      string `json:"digest"`
    Tag         string `json:"tag"`
    ResourceURL string `json:"resource_url"`
}

// Repository displays information about the repository
type Repository struct {
    DateCreated int64  `json:"date_created"`
    Name        string `json:"name"`
    Namespace   string `json:"namespace"`
    RepoFullName string `json:"repo_full_name"`
    RepoType    string `json:"repo_type"`
}

// EventData contains event details
type EventData struct {
    Resources  []Resource `json:"resources"`
    Repository Repository `json:"repository"`
}

// WebhookPayload represents the body of the webhook request
type WebhookPayload struct {
    Type     string    `json:"type"`
    OccurAt  int64     `json:"occur_at"`
    Operator string    `json:"operator"`
    EventData EventData `json:"event_data"`
}

var bot *tgbotapi.BotAPI
func initTelegramBot() {
    var err error
    bot, err = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
    if err != nil {
        log.Panic(err)
    }

    debugModeStr := os.Getenv("DEBUG_MODE")     //true - hide data, false - show data
    debugModeBool, err := strconv.ParseBool(debugModeStr)
    if err != nil {
    log.Printf("Error converting DEBUG_MODE to bool: %v", err)
    return
    }

    bot.Debug = debugModeBool
    log.Printf("Launching...")
    log.Printf("OK! Connected to telegram bot account: https://t.me/%s", bot.Self.UserName)
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

    message := fmt.Sprintf("New image pushed by: <b>%s</b>.\n", payload.Operator)
    message += fmt.Sprintf("- Host: <a href=\"%s\">%s</a>\n", harborLink, harborURL)
    message += fmt.Sprintf("- Project: <b>%s</b>\n", repo.Namespace)
    message += fmt.Sprintf("- Repository: <b>%s</b>\n", repo.RepoFullName)
    message += fmt.Sprintf("- Tag: <b>%s</b>", resource.Tag)

    return message
}

func sendTelegramMessage(chatID int64, message string) {
    msg := tgbotapi.NewMessage(chatID, message)
    msg.ParseMode = "HTML"
    if _, err := bot.Send(msg); err != nil {
        log.Println("ERROR!!! When sending message:", err)
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
    log.Printf("Error converting chat ID to int64: %v", err)
    return
    }

    // Forming and sending a message in Telegram
    message := formatMessage(payload) // Using formatMessage to create a message
    sendTelegramMessage(chatIdInt, message)

    // Response to request
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK! Webhook received."))
}

func main() {
    initTelegramBot() // Initialize the Telegram bot
    http.HandleFunc("/webhook-bot", handleWebhook)
    log.Fatal(http.ListenAndServe(":441", nil))
}
