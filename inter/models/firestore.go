package models

import (
	"cloud.google.com/go/firestore"
	"context"
)

func GetFirestoreClient(ctx context.Context, projectId string) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, projectId)
	if err != nil {
		return nil, err
	}
	return client, nil
}
