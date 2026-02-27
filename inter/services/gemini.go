package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Geminiから返却されるデータの構造
type LearningContent struct {
	SentenceKR  string `json:"sentence_kr"`
	SentenceJP  string `json:"sentence_jp"`
	Explanation string `json:"explanation"`
}

func GenerateLearningContent(ctx context.Context, word string) (*LearningContent, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-3-flash-preview")

	// Geminiへの指示（プロンプト）
	prompt := fmt.Sprintf(`
		韓国語の単語「%s」を使って、初級〜中級学習者向けの例文を1つ作成してください。
		以下のJSONフォーマットのみで回答してください。余計な解説文は一切不要です。

		{
		  "sentence_kr": "韓国語の例文",
		  "sentence_jp": "例文の日本語訳",
		  "explanation": "文法のポイントや単語の使い方の短い解説",
		}
	`, word)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		print(err.Error())
		return nil, err
	}

	// レスポンスからテキストを取り出す
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	part := resp.Candidates[0].Content.Parts[0]
	respText := fmt.Sprintf("%v", part)

	// JSONをパースして構造体に流し込む
	var content LearningContent
	// ※Geminiが ```json ... ``` のようにマーカーをつけてくる場合があるため、簡易的なクリーンアップが必要な場合があります
	err = json.Unmarshal([]byte(respText), &content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	return &content, nil
}
