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
	DocId        string                 `firestore:"doc_id" json:"docId"`
	Word         string                 `firestore:"word" json:"word"`
	Meaning      string                 `firestore:"meaning" json:"meaning"`
	ExampleKr    string                 `firestore:"example_kr" json:"example_kr"`
	ExampleJp    string                 `firestore:"example_jp" json:"example_jp"`
	AudioUrl     string                 `firestore:"audio_url" json:"audio_url"`
	IsLearned    bool                   `firestore:"is_learned" json:"is_learned"`
	PartOfSpeech string                 `firestore:"part_of_speech" json:"part_of_speech"`
	UserRef      *firestore.DocumentRef `firestore:"user_ref" json:"user_ref"`
	CreatedAt    time.Time              `firestore:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `firestore:"updated_at" json:"updated_at"`
	CreatedBy    *firestore.DocumentRef `firestore:"created_by" json:"created_by"`
	UpdatedBy    *firestore.DocumentRef `firestore:"updated_by" json:"updated_by"`
}

func CreateVocab(ctx context.Context, client *firestore.Client, vocab Vocabs) (string, error) {
	// 自動生成されるドキュメントIDを取得するため、Add() または NewDoc() を使用
	docRef := client.Collection(VocabsCollection).NewDoc()

	// 構造体にドキュメントIDをセット
	vocab.DocId = docRef.ID

	// タイムスタンプをセット
	now := time.Now()
	vocab.CreatedAt = now
	vocab.UpdatedAt = now

	// Firestoreに保存
	_, err := docRef.Set(ctx, vocab)
	if err != nil {
		return "", fmt.Errorf("failed to create vocab: %v", err)
	}

	return docRef.ID, nil
}

func DeleteVocabWithAudio(ctx context.Context, firestoreClient *firestore.Client, docId string) error {
	// 1. まずドキュメントを取得して音声URLを確認する
	docRef := firestoreClient.Collection("vocabs").Doc(docId)
	doc, err := docRef.Get(ctx)
	if err != nil {
		return fmt.Errorf("ドキュメントの取得に失敗しました: %v", err)
	}

	// 2. 音声URLを取得
	var audioURL string
	if url, ok := doc.Data()["audio_url"].(string); ok {
		audioURL = url
	}

	// 3. 音声URLがある場合、Storageからファイルを削除
	if audioURL != "" {
		// services.DeleteAudioFile は以前作成したURLからパスを抽出して消す関数
		if err := DeleteAudioFile(ctx, audioURL); err != nil {
			// ファイル削除に失敗しても、ドキュメント削除は進めるためにログに留める
			fmt.Printf("Warning: Storageの削除に失敗しました: %v\n", err)
		}
	}

	// 4. Firestoreのドキュメントを削除
	if _, err := docRef.Delete(ctx); err != nil {
		return fmt.Errorf("Firestoreの削除に失敗しました: %v", err)
	}

	return nil
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
