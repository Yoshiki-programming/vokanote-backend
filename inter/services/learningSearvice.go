package services

import (
	"cloud.google.com/go/storage"
	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	"cloud.google.com/go/texttospeech/apiv1/texttospeechpb"
	"context"
	"fmt"
)

// 音声を生成してGCSに保存し、公開URLを返す
// 引数に uid を追加
func GenerateAndUploadAudio(ctx context.Context, uid, koText, filename string) (string, error) {
	bucketName := "vokanote-audio-files"

	// 1. TTSで音声バイナリを生成
	audioBytes, err := generateSpeechBytes(ctx, koText)
	if err != nil {
		return "", err
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// 2. ストレージパスを組み立てる
	storagePath := fmt.Sprintf("users/%s/audio/%s", uid, filename)

	// オブジェクトの参照を取得
	obj := client.Bucket(bucketName).Object(storagePath)

	// 3. 書き込み（アップロード）
	sw := obj.NewWriter(ctx)
	// ブラウザが音声ファイルとして正しく認識できるように Content-Type を指定
	sw.ContentType = "audio/mpeg"

	if _, err := sw.Write(audioBytes); err != nil {
		return "", err
	}
	if err := sw.Close(); err != nil {
		return "", err
	}

	// 4. 公開権限（ACL）を設定
	// これにより、ログイン不要の公開URL（allUsers）としてアクセス可能になります
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		// 権限設定に失敗しても、URL自体は返せるのでログ出力に留める
		fmt.Printf("Warning: failed to set ACL for %s: %v\n", storagePath, err)
	}

	// 5. 公開URLを生成
	publicURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", bucketName, storagePath)
	return publicURL, nil
}

func generateSpeechBytes(ctx context.Context, koText string) ([]byte, error) {
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create tts client: %v", err)
	}
	defer client.Close()

	// 1. 韓国語
	koReq := &texttospeechpb.SynthesizeSpeechRequest{
		Input:       &texttospeechpb.SynthesisInput{InputSource: &texttospeechpb.SynthesisInput_Text{Text: koText}},
		Voice:       &texttospeechpb.VoiceSelectionParams{LanguageCode: "ko-KR", Name: "ko-KR-Neural2-A"},
		AudioConfig: &texttospeechpb.AudioConfig{AudioEncoding: texttospeechpb.AudioEncoding_MP3},
	}
	koResp, err := client.SynthesizeSpeech(ctx, koReq)
	if err != nil {
		return nil, fmt.Errorf("korean tts error: %v", err) // ここでエラーが分かります
	}

	return append(koResp.AudioContent), nil
}
