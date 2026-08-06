package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

const (
	productionTarget = "hmrkssfhcxlvjzyigufd"
	apiBase          = "http://localhost:8080"
)

type DevicePayload struct {
	LoginMethod string `json:"login_method"`
	DeviceID    string `json:"device_id"`
}

type loginRequest struct {
	UID        string        `json:"uid"`
	Credential string        `json:"credential"`
	Device     DevicePayload `json:"device"`
}

type loginResponse struct {
	Status     string `json:"status"`
	SetupToken string `json:"setup_token,omitempty"`
}

func main() {
	if err := godotenv.Load("../.env"); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: error loading .env: %v", err)
	}

	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		log.Fatal("SUPABASE_URL is not set")
	}

	fmt.Println("==================================================")
	fmt.Println("Target Database Verification")
	fmt.Println("==================================================")
	fmt.Printf("Detected Target: %s\n", supabaseURL)

	if strings.Contains(supabaseURL, productionTarget) {
		fmt.Println("\n[!] FATAL: Production database detected!")
		fmt.Println("[!] Mutating integration tests are strictly prohibited on production.")
		fmt.Println("[!] Please switch your .env to a staging project and try again.")
		os.Exit(1)
	}

	fmt.Println("\nTarget is safe. Proceeding with integration tests...")
	runTests()
}

func runTests() {
	fmt.Println("\n--- Starting Authentication Flow Test ---")
	uid := "test_integration_user_1"
	req := loginRequest{
		UID:        uid,
		Credential: "password123",
		Device: DevicePayload{
			LoginMethod: "PASSWORD",
			DeviceID:    "test_device_001",
		},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(apiBase+"/api/login", "application/json", bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Failed to call /api/login: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusUnauthorized {
		b, _ := io.ReadAll(resp.Body)
		log.Fatalf("Unexpected status: %v. Body: %s", resp.Status, string(b))
	}
	fmt.Println("✓ API responded to login attempt correctly.")
	fmt.Println("Integration tests skeleton completed successfully.")
}
