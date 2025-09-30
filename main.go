package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"io"
)

// ================= STRUCTS =================
// ==== Webhook ====
type Attributes struct {
	Details		string `json:"Details"`
}
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
	Attributes Attributes `json:"custom_attributes"`
}
// WebhookPayload represents the body of the webhook request
type WebhookPayload struct {
	Type      string    `json:"type"`
	OccurAt   int64     `json:"occur_at"`
	Operator  string    `json:"operator"`
	EventData EventData `json:"event_data"`
}
// ==== Artifact ====
// ApiTag represents the structure of the "tags" element.
type ApiTag struct {
    ID int `json:"id"`
    Immutable bool `json:"immutable"`
    Name string `json:"name"`
}
// HarborArtifact represents the main structure of the JSON response.
type HarborArtifact struct {
    Type string `json:"type"`
	ProjectId int `json:"project_id"`
    Tags []ApiTag `json:"tags"`
	// можна додати інші блоки за необхідності
}
// ==== Quota ====
type Quota struct {
    ID     int `json:"id"`
    Ref    struct {
        ID        int    `json:"id"`
        Name      string `json:"name"`
        OwnerName string `json:"owner_name"`
    } `json:"ref"`
    Hard struct {
        Storage int64 `json:"storage"`
        Count   int64 `json:"count,omitempty"` // може не бути
    } `json:"hard"`
    Used struct {
        Storage int64 `json:"storage"`
        Count   int64 `json:"count,omitempty"` // може не бути
    } `json:"used"`
    CreationTime string `json:"creation_time,omitempty"`
    UpdateTime   string `json:"update_time,omitempty"`
}
// ==== for calc Quota ====
type QuotaInfo struct {
    TotalMB   float64
    UsedMB    float64
    Percent   float64
    Warning   string // "w"-Попередження //"n"-нормас
}
// ==== For Telegram send ====
type SendMessageParams struct {
	ChatID  int64
	Message string
	TopicID *int64
}
// ================= GLOBAL =================
var bot *tgbotapi.BotAPI

var Debug bool
func init() {
	// Читаємо DEBUG з env
	debugEnv := os.Getenv("DEBUG")
	Debug = strings.ToLower(debugEnv) == "true"
}

// ================= INIT BOT =================
func initTelegramBot() {
	var err error
	bot, err = tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	debugModeStr := os.Getenv("DEBUG")
	debugModeBool, err := strconv.ParseBool(debugModeStr)
	if err != nil {
		debugModeBool = false
	}

	bot.Debug = debugModeBool
	log.Printf("OK! Connected to telegram bot account: https://t.me/%s", bot.Self.UserName)
}

// ================= HELPERS =================
func extractDomain(resourceURL string) string {
	start := strings.Index(resourceURL, "//")
	if start == -1 {
		return ""
	}

	start += 2
	end := strings.Index(resourceURL[start:], "/")
	if end == -1 {
		return resourceURL[start:]
	}

	return resourceURL[start : start+end]
}

func toJSONPretty(v interface{}) string {
	prettyJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error when marshalling to pretty JSON: %v", err)
		return ""
	}
	return string(prettyJSON)
}


// ================= HARBOR API =================
func getQuota(artifact HarborArtifact) (*Quota, error){
	hostUrl := os.Getenv("HOST")
	user := os.Getenv("HARBOR_USER")
    pass := os.Getenv("HARBOR_PASS")
	url := fmt.Sprintf(
		"%s/api/v2.0/quotas?reference=project&reference_id=%d",
		hostUrl,
		artifact.ProjectId,
	)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %v", err)
    }

    req.SetBasicAuth(user, pass)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("failed to execute request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %v", err)
    }

    var quotas []Quota
    if err := json.Unmarshal(body, &quotas); err != nil {
        return nil, fmt.Errorf("failed to parse quota JSON: %v", err)
    }

    if len(quotas) == 0 {
        return nil, fmt.Errorf("no quota info returned for project %d", artifact.ProjectId)
    }

    return &quotas[0], nil
}

func calcQuotaUsage(used, hard int64) QuotaInfo {
    if hard <= 0 {
        return QuotaInfo{
            TotalMB:  0,
            UsedMB:   float64(used) / (1024 * 1024),
            Percent:  0,
            Warning:  "n",
        }
    }

    totalMB := float64(hard) / (1024 * 1024)
    usedMB := float64(used) / (1024 * 1024)
    percent := (usedMB / totalMB) * 100

    warning := "n"
    if percent > 85 {
        warning = "w"
    }

    return QuotaInfo{TotalMB: totalMB, UsedMB:  usedMB, Percent: percent, Warning: warning, }
}

// getArtifact виконує GET-запит і повертає параметр type та tags.Name
func getArtifact(resource Resource, repo Repository) (HarborArtifact, error) {
	var artifact HarborArtifact
	hostUrl := os.Getenv("HOST") // -e HOST='http://nginx:8080'
	username := os.Getenv("HARBOR_USER") // логін
	password := os.Getenv("HARBOR_PASS") // пароль
	url := fmt.Sprintf(
		"%s/api/v2.0/projects/%s/repositories/%s/artifacts/%s",
		hostUrl, repo.Namespace, repo.Name,	resource.Digest,
	)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return artifact, fmt.Errorf("failed to create request: %v", err)
	}
	// Додаємо BasicAuth
	req.SetBasicAuth(username, password)
	// Виконуємо
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return artifact, fmt.Errorf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if Debug { //= Обмежуємо запис в лог по змінній
		log.Printf("\nDEBUG: raw API_RESPONSE body:\n %s\n", string(body))
	}

	if err != nil {
		log.Printf("ERROR: failed to read response body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}	
		
	if err := json.Unmarshal(body, &artifact); err != nil {
		log.Printf("ERROR: failed to unmarshal artifact response: %v", err)
	}

	if Debug { log.Printf("\nDEBUG: Artifact type: %s, tag: %v, quota: %d", artifact.Type, artifact.Tags[0].Name, artifact.ProjectId) }
	return artifact, nil
}

// ================= TELEGRAM =================

func formatMessage(payload WebhookPayload, artifact HarborArtifact, qu QuotaInfo, hostUrl string) string {
	var (
    resource   Resource
    repo       Repository
    harborURL  string
    harborLink string
	)
	switch payload.Type {
	case "PUSH_ARTIFACT", "PULL_ARTIFACT", "DELETE_ARTIFACT", "QUOTA_WARNING":
		if len(payload.EventData.Resources) == 0 {
			log.Printf("&#9888; Webhook Event Received, but no resources are available.")
			return ""
		}
		resource = payload.EventData.Resources[0]
		repo = payload.EventData.Repository
		harborURL = strings.Split(resource.ResourceURL, "/")[0]
		harborLink = fmt.Sprintf("https://%s/harbor/projects", harborURL)
	}

	if Debug { log.Printf("DEBUG in formatMessage: artifact=%+v", artifact) }
	//log.Printf("\nCheck into formatMessage: Artifact type: %s, tag: %v", artifactType, artifactTag)

	var message string
	switch payload.Type {
	case "PUSH_ARTIFACT":
		if artifact.Type == "IMAGE" {
		message = fmt.Sprintf("&#128051; New image pushed by: <b>%s</b>\n", payload.Operator)
		} else if artifact.Type == "CHART" {
		message = fmt.Sprintf("&#9784; New chart pushed by: <b>%s</b>\n", payload.Operator)
		} else {
		message = fmt.Sprintf("&#128230; New artifact pushed by: <b>%s</b>\n", payload.Operator)
		}
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", harborLink, harborURL)
		message += fmt.Sprintf("• Access: <b>%s</b>\n", repo.RepoType)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", repo.Namespace)
		message += fmt.Sprintf("• Repository: <b>%s</b>\n", repo.RepoFullName)
		message += fmt.Sprintf("• Tag: <b>%s</b>\n", resource.Tag)
		if qu.Warning == "w" {
			message += "\n&#9888; Warning!! Quota usage reach 85%!!\n"
			message += fmt.Sprintf("• Details: <b>quota usage reach %.2f%%: resource storage used %.2f MB of %.2f MB</b>\n", qu.Percent, qu.UsedMB, qu.TotalMB)
		}
	case "PULL_ARTIFACT":
		if artifact.Type == "IMAGE" {
		message = fmt.Sprintf("&#128051; Image pulled by: <b>%s</b>\n", payload.Operator)
		} else if artifact.Type == "CHART" {
		message = fmt.Sprintf("&#9784; Chart pulled by: <b>%s</b>\n", payload.Operator)
		} else {
		message = fmt.Sprintf("&#128230; Artifact pulled by: <b>%s</b>\n", payload.Operator)
		}
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", harborLink, harborURL)
		message += fmt.Sprintf("• Access: <b>%s</b>\n", repo.RepoType)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", repo.Namespace)
		message += fmt.Sprintf("• Repository: <b>%s</b>\n", repo.RepoFullName)
		message += fmt.Sprintf("• Tag: <b>%s</b>", artifact.Tags[0].Name)
	case "DELETE_ARTIFACT":
		if artifact.Type == "IMAGE" {
		message = fmt.Sprintf("&#10071; Attention!\n&#128051; Image removed by: <b>%s</b>\n", payload.Operator)
		} else if artifact.Type == "CHART" {
		message = fmt.Sprintf("&#10071; Attention!\n&#9784; Chart removed by: <b>%s</b>\n", payload.Operator)
		} else {
		message = fmt.Sprintf("&#10071; Attention!\n&#128230; Artifact removed by: <b>%s</b>\n", payload.Operator)
		}
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", harborLink, harborURL)
		message += fmt.Sprintf("• Access: <b>%s</b>\n", repo.RepoType)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", repo.Namespace)
		message += fmt.Sprintf("• Repository: <b>%s</b>\n", repo.RepoFullName)
		message += fmt.Sprintf("• Tag: <b>%s</b>", resource.Tag)
	case "QUOTA_WARNING":
		message = "&#9888; Warning!! Quota usage reach 85%!!\n"
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", fmt.Sprintf("%s/harbor/projects", hostUrl), hostUrl)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", payload.EventData.Repository.Namespace)
		message += fmt.Sprintf("• Details: <b>%s</b>\n", payload.EventData.Attributes.Details)
	case "QUOTA_EXCEED":
		message = "&#128680; Alert!!! Project quota has been exceed!!!\n"
		message += fmt.Sprintf("• Host: <a href=\"%s\">%s</a>\n", fmt.Sprintf("%s/harbor/projects", hostUrl), hostUrl)
		message += fmt.Sprintf("• Project: <b>%s</b>\n", payload.EventData.Repository.Namespace)
		message += fmt.Sprintf("• Details: <b>%s</b>\n", payload.EventData.Attributes.Details)
	default:
		message = "&#9888; Received an unknown event type."
	}

	return message
}

func sendTelegramMessage(params SendMessageParams) {
	msg := tgbotapi.NewMessage(params.ChatID, params.Message)
	if params.TopicID != nil {
		msg.ReplyToMessageID = int(*params.TopicID)
	}
	msg.ParseMode = "HTML"

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Error when sending message: %v", err)
	}
}

// ================= HANDLER =================
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "ERROR!!! Only POST method is allowed.", http.StatusMethodNotAllowed)
		return
	}

	var payload WebhookPayload
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// логування HTTP_REQUEST від HarborWebHook
	if Debug {	log.Printf("\nDEBUG: raw HTTP_REQUEST body:\n %s\n", decoder) }

	chatIdStr := os.Getenv("CHAT_ID")
	chatIdInt, err := strconv.ParseInt(chatIdStr, 10, 64)
	if err != nil {
		log.Printf("ERROR!!! When converting chat ID: %v", err)
		return
	}

	topicIDStr := os.Getenv("TOPIC_ID")
	var topicIDPtr *int64
	if topicIDStr != "" {
		topicID, err := strconv.ParseInt(topicIDStr, 10, 64)
		if err == nil {
			topicIDPtr = &topicID
		}
	}

	// Отримуємо type з Harbor API
	var artifact HarborArtifact
	var qu QuotaInfo
	hostUrl := os.Getenv("HOST")
	switch payload.Type {
	case "PUSH_ARTIFACT", "PULL_ARTIFACT"://, "DELETE_ARTIFACT", "QUOTA_EXCEED", "QUOTA_WARNING":
		if len(payload.EventData.Resources) > 0 {
			resource := payload.EventData.Resources[0]
			repo := payload.EventData.Repository

			artifact, err = getArtifact(resource, repo)
			if err != nil {
				log.Printf("ERROR!!! Failed to get artifact type: %v", err)
			}
			//==
			quota, err := getQuota(artifact)
			if err != nil {
				log.Printf("Error getting quota: %v", err)
				return
			}
			qu = calcQuotaUsage(quota.Used.Storage, quota.Hard.Storage)
			if Debug { fmt.Printf("DEBUG: Quota Total: %.2f MB, Used: %.2f MB, Percent: %.2f%%, Flag: %s\n", qu.TotalMB, qu.UsedMB, qu.Percent, qu.Warning)}
			//==
		}
	default:
		log.Printf("Skipping event type: %s", payload.Type)
	}

	// Формуємо повідомлення
	message := formatMessage(payload, artifact, qu, hostUrl)
	sendTelegramMessage(SendMessageParams{
		ChatID:  chatIdInt,
		Message: message,
		TopicID: topicIDPtr,
	})

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK! Webhook received."))
}

// ================= MAIN =================
func main() {
	initTelegramBot()
	http.HandleFunc("/webhook-bot", handleWebhook)
	log.Fatal(http.ListenAndServe(":441", nil))
}
