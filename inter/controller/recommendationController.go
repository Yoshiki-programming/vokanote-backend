package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"google.golang.org/genai"
)

type GeminiRecommendationResponse struct {
	ExampleKr     string `json:"example_kr"`
	ExampleJp     string `json:"example_jp"`
	SongTitle     string `json:"song_title"`
	ArtistName    string `json:"artist_name"`
	LyricsSnippet string `json:"lyrics_snippet"`
	Reason        string `json:"reason"`
}

func GenerateSongRecommendation(ctx context.Context, geminiClient *genai.Client, word string) (*GeminiRecommendationResponse, error) {
	prompt := fmt.Sprintf(`
以下の韓国語の単語について情報を生成し、指定したJSON形式で出力してください。
単語: "%s"

出力形式:
{
  "example_kr": "その単語を使った韓国語の例文",
  "example_jp": "例文の日本語訳",
  "song_title": "その単語のニュアンスに合うK-POPの曲名",
  "artist_name": "アーティスト名",
  "lyrics_snippet": "その単語が含まれる、または関連する歌詞のフレーズ",
  "reason": "なぜこの曲を選んだのかという短い理由"
}
`, word)

	// SDKのバージョンにより、Configのフィールド名が ResponseMIMEType になっている場合があります
	result, err := geminiClient.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %v", err)
	}

	var resp GeminiRecommendationResponse
	// result.Text() でAIの回答を取得
	if err := json.Unmarshal([]byte(result.Text()), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse gemini response: %v", err)
	}

	return &resp, nil
}
