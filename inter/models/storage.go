package models

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/iterator"
	"strings"
)

func DeleteAudioFile(ctx context.Context, audioURL string) error {
	bucketName := "vokanote-audio-files"
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// URLからドメインとバケット名の部分を削ってパスを抽出
	// 例: https://storage.googleapis.com/vokanote-audio-files/users/123/audio/word.mp3
	// -> users/123/audio/word.mp3
	prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", bucketName)
	storagePath := strings.TrimPrefix(audioURL, prefix)

	// ファイルを削除
	err = client.Bucket(bucketName).Object(storagePath).Delete(ctx)
	if err != nil && err != storage.ErrObjectNotExist {
		return err
	}

	return nil
}

func DeleteAllAudioFile(ctx context.Context, uid string) error {
	bucketName := "vokanote-audio-files"
	client, _ := storage.NewClient(ctx)
	bucket := client.Bucket(bucketName)

	// 1. ユーザー専用フォルダ以下のファイルをすべてリストアップ
	prefix := fmt.Sprintf("users/%s/", uid)
	it := bucket.Objects(ctx, &storage.Query{Prefix: prefix})

	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		// 2. 見つかったファイルを順次削除
		bucket.Object(attrs.Name).Delete(ctx)
	}

	// 3. その後、Firestore側の全ドキュメントを削除する処理へ...
	return nil
}
