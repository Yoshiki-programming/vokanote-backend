package models

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"time"
)

const VocabsCollection = "vocabs"

type Vocabs struct {
	DocId        string                 `json:"docId"` // ドキュメント固有のID
	Word         string                 `json:"word"`
	Meaning      string                 `json:"meaning"`
	ExampleKr    string                 `json:"example_kr"`
	ExampleJp    string                 `json:"example_jp"`
	IsLearned    bool                   `json:"is_learned"`
	PartOfSpeech string                 `json:"part_of_speech"`
	UserRef      *firestore.DocumentRef `json:"user_ref"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	CreatedBy    *firestore.DocumentRef `json:"created_by"`
	UpdatedBy    *firestore.DocumentRef `json:"updated_by"`
}

func DeleteAllUserVocabs(userRef *firestore.DocumentRef, ctx context.Context, firestoreClient *firestore.Client) error {
	// 1. user_ref が一致するドキュメントを検索
	iter := firestoreClient.Collection(VocabsCollection).Where("user_ref", "==", userRef).Documents(ctx)

	batch := firestoreClient.Batch()
	hasDocs := false

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating vocabs: %v", err)
		}

		batch.Delete(doc.Ref)
		hasDocs = true
	}

	if hasDocs {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit batch delete: %v", err)
		}
	}

	return nil
}
