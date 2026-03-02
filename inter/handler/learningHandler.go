package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yoshiki-programming/vokanote-backend/inter/models"
	"github.com/Yoshiki-programming/vokanote-backend/inter/responses"
	"github.com/Yoshiki-programming/vokanote-backend/inter/services"
	"net/http"
	"os"
	"strings"
	"time"
)

func GenerateContentHandler(w http.ResponseWriter, r *http.Request) {
	// ログの開始マーカー
	fmt.Printf("[DEBUG] GenerateContentHandler started: %s %s\n", r.Method, r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ctx := context.Background()

	// 1. 環境変数の確認
	projectID := os.Getenv("PROJECT_ID")
	fmt.Printf("[DEBUG] Project ID from env: %s\n", projectID)

	firestoreClient, err := models.GetFirestoreClient(ctx, projectID)
	if err != nil {
		fmt.Printf("[ERROR] Firestore init failed: %v\n", err)
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer firestoreClient.Close()

	// 2. 認証ログ
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		fmt.Println("[ERROR] Authorization header is missing")
		responses.SendErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("missing auth header"))
		return
	}

	idToken := strings.Replace(authHeader, "Bearer ", "", 1)
	authClient, err := models.GetAuthClient(ctx) // エラーを無視せずチェック
	if err != nil {
		fmt.Printf("[ERROR] GetAuthClient failed: %v\n", err)
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		fmt.Printf("[ERROR] Token verification failed: %v\n", err)
		responses.SendErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	uid := token.UID
	userRef := firestoreClient.Collection("users").Doc(uid)
	fmt.Printf("[DEBUG] Auth Success: User UID = %s\n", uid)

	// 3. 🔥 リクエストパラメータ解析（JSON対応に修正）
	var reqBody struct {
		Word string `json:"word"`
		Pos  string `json:"part_of_speech"`
	}

	// JSONデコードを試みる
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		fmt.Printf("[ERROR] JSON Decode failed: %v. Checking FormValue instead.\n", err)
		// 念のためFormValueもチェック（フォールバック）
		reqBody.Word = r.FormValue("word")
		reqBody.Pos = r.FormValue("part_of_speech")
	}

	fmt.Printf("[DEBUG] Received Input: word=[%s], pos=[%s]\n", reqBody.Word, reqBody.Pos)

	if reqBody.Word == "" || reqBody.Pos == "" {
		fmt.Println("[ERROR] Validation Failed: word or pos is empty")
		responses.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("word and part_of_speech are required"))
		return
	}

	// 4. 重複チェックログ
	fmt.Printf("[DEBUG] Checking cache for: %s\n", reqBody.Word)
	iter := firestoreClient.Collection("vocabs").
		Where("user_ref", "==", userRef).
		Where("word", "==", reqBody.Word).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == nil {
		fmt.Printf("[CACHE] Hit: %s (Skip Gemini/TTS)\n", reqBody.Word)
		finalResponse := map[string]interface{}{
			"docId":   doc.Ref.ID,
			"message": "already exists",
		}
		resBody, _ := json.Marshal(finalResponse)
		responses.SendResponse(w, http.StatusOK, resBody)
		return
	}

	// 5. 生成ロジック開始ログ
	fmt.Printf("[DEBUG] Calling services.GenerateLearningContent for: %s\n", reqBody.Word)
	content, err := services.GenerateLearningContent(ctx, reqBody.Word)
	if err != nil {
		fmt.Printf("[ERROR] Gemini API/Service failed: %v\n", err)
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	fmt.Printf("[DEBUG] TTS Generating for: %s\n", content.SentenceKR)
	fileName := fmt.Sprintf("%s_%d.mp3", reqBody.Word, time.Now().Unix())
	audioURL, err := services.GenerateAndUploadAudio(ctx, uid, content.SentenceKR, fileName)
	if err != nil {
		fmt.Printf("[ERROR] TTS Error: %v. Continuing without audio.\n", err)
		audioURL = ""
	}

	// 6. 最終書き込みログ
	fmt.Printf("[DEBUG] Creating Firestore doc for word: %s\n", reqBody.Word)
	newVocab := models.Vocabs{
		Word:         reqBody.Word,
		Meaning:      content.Explanation,
		ExampleKr:    content.SentenceKR,
		ExampleJp:    content.SentenceJP,
		AudioUrl:     audioURL,
		IsLearned:    false,
		PartOfSpeech: reqBody.Pos,
		UserRef:      userRef,
		CreatedBy:    userRef,
		UpdatedBy:    userRef,
	}

	docId, err := models.CreateVocab(ctx, firestoreClient, newVocab)
	if err != nil {
		fmt.Printf("[ERROR] Firestore CreateVocab failed: %v\n", err)
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	fmt.Printf("[SUCCESS] Process completed for: %s, docId: %s\n", reqBody.Word, docId)

	finalResponse := map[string]interface{}{
		"docId":  docId,
		"status": "success",
	}
	resBody, _ := json.Marshal(finalResponse)
	responses.SendResponse(w, http.StatusOK, resBody)
}
