package models

import (
	"cloud.google.com/go/firestore"
	"context"
	"google.golang.org/api/option"
)

func GetFirestoreClient(ctx context.Context, opt option.ClientOption, projectId string) (*firestore.Client, error) {
	client, err := firestore.NewClient(ctx, projectId, opt)
	if err != nil {
		return nil, err
	}
	return client, err
}
