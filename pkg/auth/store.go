package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// firestoreStore implements FirestoreReader using cloud.google.com/go/firestore.
// This is the concrete infrastructure adapter — unit tests use a mock
// FirestoreReader instead, so this file need not be exercised in `go test`
// without a live Firestore instance.
type firestoreStore struct {
	client *firestore.Client
}

// NewFirestoreStore wraps an existing *firestore.Client as a FirestoreReader.
func NewFirestoreStore(client *firestore.Client) FirestoreReader {
	return &firestoreStore{client: client}
}

// GetDeviceDocument reads the Gatekeeper device document at
// users/{parentID}/children/{uid} (read-only).
func (s *firestoreStore) GetDeviceDocument(ctx context.Context, parentID, uid string) (map[string]any, error) {
	path := fmt.Sprintf("users/%s/children/%s", parentID, uid)
	docSnap, err := s.client.Doc(path).Get(ctx)
	if err != nil {
		return nil, err
	}
	return docSnap.Data(), nil
}

// NewFirestoreClient creates a Firestore client from environment credentials.
// Supports raw JSON via GOOGLE_APPLICATION_CREDENTIALS_JSON or FIREBASE_CREDENTIALS
// (with automatic base64 decoding as a fallback for transport-escaped values).
func NewFirestoreClient(ctx context.Context) (*firestore.Client, error) {
	credJSON := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS_JSON")
	if credJSON == "" {
		credJSON = os.Getenv("FIREBASE_CREDENTIALS")
	}
	if credJSON == "" {
		return nil, fmt.Errorf("firestore credentials not found: set GOOGLE_APPLICATION_CREDENTIALS_JSON or FIREBASE_CREDENTIALS")
	}

	credJSON = resolveCredentialsJSON(credJSON)

	var projectID string
	var fullMap map[string]any
	if err := json.Unmarshal([]byte(credJSON), &fullMap); err != nil {
		return nil, fmt.Errorf("invalid credentials JSON: %w", err)
	}
	if pid, ok := fullMap["project_id"].(string); ok {
		projectID = pid
	}
	if projectID == "" {
		return nil, fmt.Errorf("project_id missing in credentials")
	}

	// Fix escaped newlines in private_key (Vercel-style single-line JSON).
	if rawKey, ok := fullMap["private_key"].(string); ok {
		fullMap["private_key"] = strings.ReplaceAll(rawKey, "\\n", "\n")
	}
	fixedCreds, err := json.Marshal(fullMap)
	if err != nil {
		return nil, fmt.Errorf("re-marshal credentials: %w", err)
	}

	client, err := firestore.NewClient(ctx, projectID, option.WithCredentialsJSON(fixedCreds))
	if err != nil {
		return nil, fmt.Errorf("create firestore client: %w", err)
	}
	return client, nil
}

// resolveCredentialsJSON handles base64-encoded credential payloads that may
// arrive when environment variables are transport-escaped.
func resolveCredentialsJSON(credJSON string) string {
	credJSON = strings.TrimSpace(credJSON)
	if json.Valid([]byte(credJSON)) {
		return credJSON
	}
	if decoded, err := base64.StdEncoding.DecodeString(credJSON); err == nil {
		if json.Valid(decoded) {
			return string(decoded)
		}
	}
	return credJSON
}
