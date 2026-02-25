package models

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"time"
)

const RecommendationsCollection = "recommendations"

type Recommendations struct {
	DocId          string                 `json:"docId"`
	VocabRef       *firestore.DocumentRef `json:"vocab_ref"`
	SongTitle      string                 `json:"song_title"`
	ArtistName     string                 `json:"artist_name"`
	YoutubeVideoId string                 `json:"youtube_video_id"`
	ThumbnailUrl   string                 `json:"thumbnail_url"`
	LyricsSnippet  string                 `json:"lyrics_snippet"`
	StartAt        int                    `json:"start_at"` // 秒単位の再生開始位置
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedBy      *firestore.DocumentRef `json:"created_by"`
}

func DeleteAllUserRecommendations(userRef *firestore.DocumentRef, ctx context.Context, firestoreClient *firestore.Client) error {
	// 1. CreatedBy が一致するドキュメントを検索
	iter := firestoreClient.Collection(RecommendationsCollection).Where("created_by", "==", userRef).Documents(ctx)

	batch := firestoreClient.Batch()
	hasDocs := false

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating recommendations: %v", err)
		}

		// 削除対象をバッチに追加
		batch.Delete(doc.Ref)
		hasDocs = true
	}

	// 2. 1件以上あればコミット
	if hasDocs {
		if _, err := batch.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit batch delete for recommendations: %v", err)
		}
	}

	return nil
}
