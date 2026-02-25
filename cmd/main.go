package main

import (
	"fmt"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/utils/Slog"
	"github.com/Yoshiki-programming/vokanote-backend.git/router"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

// main関数は、アプリケーションのエントリーポイントです
func main() {

	// 環境変数読み込み
	err := godotenv.Load(fmt.Sprintf("configs/%s.env", os.Getenv("GO_ENV")))
	if err != nil {
		log.Fatalf("failed to load env file: %v", err)
	}

	//ルーティング
	router.Router()

	//PORT番号を.envから取得
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		Slog.Warn("port number is not set. default: 8080")
	}

	Slog.DebugInfo("PROJECT_ID", os.Getenv("PROJECT_ID"))

	// サーバーの起動
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("server startup failed: %v", err)
	}
}
