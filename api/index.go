package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
)

var db *sql.DB

func main() {
	// Inisialisasi koneksi database
	var err error
	db, err = sql.Open("mysql", "mielola:140804@tcp(localhost:3306)/github-notif")
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		return
	}
	defer db.Close()

	// Buat instance Echo
	e := echo.New()

	// Route untuk webhook GitHub
	e.POST("/webhook", handleWebhook)

	// Mulai server
	e.Start(":8080")
}

func handleWebhook(c echo.Context) error {
	fmt.Println("Received webhook")

	// Log request headers
	for k, v := range c.Request().Header {
		fmt.Printf("Header %s: %v\n", k, v)
	}

	// Log request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Println("Error reading request body:", err)
		return c.String(http.StatusInternalServerError, "Error reading request body")
	}
	fmt.Println("Request body:", string(body))
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	// Parse payload webhook
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		fmt.Println("Error parsing JSON payload:", err)
		return c.String(http.StatusBadRequest, "Invalid JSON payload")
	}

	// Proses payload
	event := c.Request().Header.Get("X-GitHub-Event")
	switch event {
	case "push":
		handlePushEvent(payload)
	// Tambahkan case lain untuk event yang berbeda
	default:
		fmt.Println("Unhandled event:", event)
	}

	return c.String(http.StatusOK, "Webhook received")
}

func handlePushEvent(payload map[string]interface{}) {
	if payload == nil {
		fmt.Println("Error: Payload is nil")
		return
	}

	repo, ok := payload["repository"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Unable to parse repository information")
		return
	}

	repoName, ok := repo["full_name"].(string)
	if !ok {
		fmt.Println("Error: Unable to parse repository name")
		return
	}

	pusher, ok := payload["pusher"].(map[string]interface{})
	if !ok {
		fmt.Println("Error: Unable to parse pusher information")
		return
	}

	pusherName, ok := pusher["name"].(string)
	if !ok {
		fmt.Println("Error: Unable to parse pusher name")
		return
	}

	commits, ok := payload["commits"].([]interface{})
	if !ok {
		fmt.Println("Error: Unable to parse commits")
		return
	}

	var messageBuilder strings.Builder
	messageBuilder.WriteString(fmt.Sprintf("Ada yang baru pushh nihhh ke github %s by %s\n\n", repoName, pusherName))

	for _, c := range commits {
		commit, ok := c.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := commit["id"].(string)
		message, _ := commit["message"].(string)
		timestamp, _ := commit["timestamp"].(string)
		author, _ := commit["author"].(map[string]interface{})
		authorName, _ := author["name"].(string)

		messageBuilder.WriteString(fmt.Sprintf("Commit: %s\n", id[:7]))
		messageBuilder.WriteString(fmt.Sprintf("Author: %s\n", authorName))
		messageBuilder.WriteString(fmt.Sprintf("Date: %s\n", timestamp))
		messageBuilder.WriteString(fmt.Sprintf("Message: %s\n\n", message))
	}

	// Simpan informasi ke database
	_, err := db.Exec("INSERT INTO push_events (repo_name, pusher_name) VALUES (?, ?)", repoName, pusherName)
	if err != nil {
		fmt.Println("Error inserting into database:", err)
	}

	// Kirim pesan ke bot Telegram
	sendTelegramMessage(messageBuilder.String())
}

func sendTelegramMessage(message string) {
	botToken := "7864060958:AAGvKAtDiNb9GCr6JaaZ7Vnh67gAk5k67cQ"
	chatID := "-4521278263"

	if botToken == "" || chatID == "" {
		fmt.Println("Error: TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is not set")
		return
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	fmt.Println("Sending message to Telegram...")
	fmt.Println("URL:", url)
	fmt.Println("Chat ID:", chatID)
	fmt.Println("Message:", message)

	resp, err := http.PostForm(url, map[string][]string{
		"chat_id": {chatID},
		"text":    {message},
	})
	if err != nil {
		fmt.Println("Error sending Telegram message:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Telegram API response:", string(body))

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status:", resp.StatusCode)
	}
}
