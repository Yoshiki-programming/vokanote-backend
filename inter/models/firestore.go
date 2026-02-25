package models

import (
	"cloud.google.com/go/firestore"
	"context"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/utils/Slog"
	"google.golang.org/api/option"
	"os"
)

func GetFirestoreClient(ctx context.Context, opt option.ClientOption, projectId string) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, projectId, opt)
	if err != nil {
		return nil, err
	}
	return client, err
}

// GetFirebaseAuth FirebaseAuthClientを返す
func GetFirebaseAuth() (*auth.Client, error) {
	ctx := context.Background()
	opt := option.WithCredentialsJSON([]byte(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")))
	config := &firebase.Config{ProjectID: os.Getenv("PROJECT_ID")}
	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		Slog.DebugInfo("error initializing app: %v\n", err)
		return nil, err
	}

	// Auth Client
	client, err := app.Auth(ctx)
	if err != nil {
		Slog.DebugInfo("error getting Auth client: %v\n", err)
		return nil, err
	}
	return client, err
}
