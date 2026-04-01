//go:build ignore

package main

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Automatically load the .env file if it exists in the root directory
	_ = godotenv.Load()

	// Define command-line flags
	clientID := flag.String("id", "", "The client ID (e.g., 'media-service')")
	name := flag.String("name", "", "A human-readable name for the client")
	useVault := flag.Bool("vault", false, "Push the raw secret directly to HashiCorp Vault")
	flag.Parse()

	// Validate required flags
	if *clientID == "" || *name == "" {
		log.Fatal("❌ Usage: go run scripts/create_m2m_client.go -id <id> -name <name> [-vault]")
	}

	// 1. Generate 32 bytes of secure randomness for the Secret
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("❌ Failed to generate random secret: %v", err)
	}
	secret := hex.EncodeToString(b)

	// 2. Hash the secret using bcrypt (for the Auth Database)
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ Failed to hash secret: %v", err)
	}
	hash := string(hashBytes)

	// 3. Database Transaction Setup
	dbURL := os.Getenv("DATABASE_URL")

	// Automatically fallback to localhost when running the script outside the Docker network.
	dbURL = strings.Replace(dbURL, "auth-storage-timescale", "localhost", 1)

	var db *sql.DB
	var tx *sql.Tx
	useDB := dbURL != ""

	if useDB {
		var err error
		db, err = sql.Open("pgx", dbURL)
		if err != nil {
			log.Fatalf("❌ Failed to connect to database: %v", err)
		}
		defer db.Close()

		tx, err = db.Begin()
		if err != nil {
			log.Fatalf("❌ Failed to begin database transaction: %v", err)
		}
		defer tx.Rollback() // Rollback is a no-op if transaction is already committed

		query := `INSERT INTO machine_clients (client_id, client_secret_hash, name, scopes) VALUES ($1, $2, $3, '{}')`
		_, err = tx.Exec(query, *clientID, hash, *name)
		if err != nil {
			log.Fatalf("❌ Failed to insert machine client into database: %v", err)
		}
	}

	// 4. Handle the Raw Secret (Vault vs. Console)
	if *useVault {
		vaultAddr := os.Getenv("VAULT_ADDR")
		vaultToken := os.Getenv("VAULT_DEV_ROOT_TOKEN_ID")

		// Fallback just in case you use a standard VAULT_TOKEN env var
		if vaultToken == "" {
			vaultToken = os.Getenv("VAULT_TOKEN")
		}

		if vaultAddr == "" || vaultToken == "" {
			log.Fatal("❌ VAULT_ADDR and VAULT_DEV_ROOT_TOKEN_ID must be set in your .env file to use Vault")
		}

		// Prepare the JSON payload for Vault KV-V2 Engine
		payload := map[string]any{
			"data": map[string]string{
				"client_id":     *clientID,
				"client_secret": secret,
			},
		}
		jsonData, _ := json.Marshal(payload)

		// Vault API Endpoint path
		url := fmt.Sprintf("%s/v1/secret/data/storage/m2m/%s", vaultAddr, *clientID)
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		req.Header.Set("X-Vault-Token", vaultToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("❌ Vault Push Failed (Network Error): %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			log.Fatalf("❌ Vault Push Failed. HTTP Status: %d. Did you run 'make init-vault'?", resp.StatusCode)
		}

		fmt.Printf("✅ Secret securely stored in Vault at: secret/storage/m2m/%s\n", *clientID)
	} else {
		// If NOT using vault, print the secret so the developer can copy it
		fmt.Println("==================================================")
		fmt.Printf("✅ Machine Client Credentials Generated!\n\n")
		fmt.Printf("Client ID     : %s\n", *clientID)
		fmt.Printf("Client Secret : %s\n", secret)
		fmt.Println("==================================================")
		fmt.Println("⚠️  SAVE THE SECRET ABOVE. IT WILL NEVER BE SHOWN AGAIN.")
	}

	// 5. Commit Database Transaction
	if useDB {
		if err := tx.Commit(); err != nil {
			log.Fatalf("❌ Failed to commit database transaction: %v", err)
		}
		fmt.Printf("✅ Machine client '%s' successfully inserted into the database!\n", *clientID)
	} else {
		log.Println("⚠️ DATABASE_URL not set. Falling back to printing SQL statement.")
		fmt.Println("\n--- SQL FOR AUTH DATABASE ---")
		fmt.Printf("INSERT INTO machine_clients (client_id, client_secret_hash, name, scopes) \nVALUES ('%s', '%s', '%s', '{}');\n", *clientID, hash, *name)
	}
}
