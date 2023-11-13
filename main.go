package main

import (
    "encoding/json"
    "log"
    "net/http"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Resource представляє ресурс, пов'язаний з подією
type Resource struct {
    Digest      string `json:"digest"`
    Tag         string `json:"tag"`
    ResourceURL string `json:"resource_url"`
}

// Repository відображає інформацію про репозиторій
type Repository struct {
    DateCreated int64  `json:"date_created"`
    Name        string `json:"name"`
    Namespace   string `json:"namespace"`
    RepoFullName string `json:"repo_full_name"`
    RepoType    string `json:"repo_type"`
}

// EventData містить деталі події
type EventData struct {
    Resources  []Resource `json:"resources"`
    Repository Repository `json:"repository"`
}

// WebhookPayload представляє тіло webhook запиту
type WebhookPayload struct {
    Type     string    `json:"type"`
    OccurAt  int64     `json:"occur_at"`
    Operator string    `json:"operator"`
    EventData EventData `json:"event_data"`
}

func handleWebhook(w http.ResponseWriter, r *http.Request) {
    // Перевірка на POST запит
    if r.Method != "POST" {
        http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
        return
    }

    // Декодування JSON з тіла запиту
    var payload WebhookPayload
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&payload); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer r.Body.Close()

    // Логування отриманих даних
    log.Printf("Received PUSH_ARTIFACT event: %+v\n", payload)

    // Відповідь на запит
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Webhook received"))
}

func main() {
    http.HandleFunc("/webhook-bot", handleWebhook)
    log.Fatal(http.ListenAndServe(":441", nil))
}
