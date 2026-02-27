package models

import (
	"context"

	"google.golang.org/genai"
)

// Geminiクライアントを初期化する関数
func GetGeminiClient(ctx context.Context) (*genai.Client, error) {
	// GEMINI_API_KEY は環境変数から自動で読み込まれます
	// 第2引数の Config で API キーを明示的に指定することも可能です
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}
