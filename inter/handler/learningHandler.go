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
	// 1. CORS & Method チェック (省略)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	ctx := context.Background()

	// 2. クライアント & 認証 (省略)
	firestoreClient, err := models.GetFirestoreClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer firestoreClient.Close()

	authHeader := r.Header.Get("Authorization")
	idToken := strings.Replace(authHeader, "Bearer ", "", 1)
	authClient, _ := models.GetAuthClient(ctx)
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusUnauthorized, err)
		return
	}
	uid := token.UID
	userRef := firestoreClient.Collection("users").Doc(uid)

	// 3. リクエストパラメータ解析
	word := r.FormValue("word")
	pos := r.FormValue("part_of_speech")
	if word == "" {
		responses.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("word is required"))
		return
	}
	if pos == "" {
		responses.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("part of speech is required"))
		return
	}

	// 4. 🔥 重複チェック (節約ロジック)
	iter := firestoreClient.Collection("vocabs").
		Where("user_ref", "==", userRef).
		Where("word", "==", word).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err == nil {
		// すでに登録済みなら、保存されているデータをそのまま返す
		var vocab models.Vocabs
		doc.DataTo(&vocab)

		fmt.Printf("Cache Hit: %s (Skip Gemini/TTS)\n", word)

		finalResponse := map[string]interface{}{
			"docId":   doc.Ref.ID,
			"message": "already exists",
		}
		resBody, _ := json.Marshal(finalResponse)
		responses.SendResponse(w, http.StatusOK, resBody)
		return
	}

	// 5. 新規生成ロジック (ここから先は「未登録」の場合のみ実行される)
	content, err := services.GenerateLearningContent(ctx, word)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	fileName := fmt.Sprintf("%s_%d.mp3", word, time.Now().Unix())
	audioURL, err := services.GenerateAndUploadAudio(ctx, uid, content.SentenceKR, fileName)
	if err != nil {
		fmt.Printf("TTS Error: %v\n", err)
		audioURL = ""
	}

	newVocab := models.Vocabs{
		Word:         word,
		Meaning:      content.Explanation,
		ExampleKr:    content.SentenceKR,
		ExampleJp:    content.SentenceJP,
		AudioUrl:     audioURL,
		IsLearned:    false,
		PartOfSpeech: pos,
		UserRef:      userRef,
		CreatedBy:    userRef,
		UpdatedBy:    userRef,
	}

	docId, err := models.CreateVocab(ctx, firestoreClient, newVocab)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// キャッシュヒット時も新規作成時も、これだけで十分
	finalResponse := map[string]interface{}{
		"docId":  docId, // FF側でこのIDを使ってページ遷移や更新をするため
		"status": "success",
	}
	resBody, _ := json.Marshal(finalResponse)
	responses.SendResponse(w, http.StatusOK, resBody)
}
