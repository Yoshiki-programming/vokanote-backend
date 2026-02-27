package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Yoshiki-programming/vokanote-backend/inter/models"
	"github.com/Yoshiki-programming/vokanote-backend/inter/responses"
	"net/http"
	"os"
	"strings"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	// 1. CORS ヘッダー
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if r.Method != http.MethodPost {
		responses.SendErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("%s method not allowed", r.Method))
		return
	}

	ctx := context.Background()

	// 2. クライアント初期化
	firestoreClient, err := models.GetFirestoreClient(ctx, os.Getenv("PROJECT_ID"))
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer firestoreClient.Close()

	authClient, err := models.GetAuthClient(ctx)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// 3. 認証
	authHeader := r.Header.Get("Authorization")
	idToken := strings.Replace(authHeader, "Bearer ", "", 1)
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}
	uid := token.UID

	// 4. パラメータ取得
	mode := r.FormValue("mode") // "single" or "all"
	docId := r.FormValue("docId")

	// 5. モード別処理
	if mode == "single" {
		// --- 【個別削除】 ---
		if docId == "" {
			responses.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("docId is required for single mode"))
			return
		}
		// Firestore内URL確認 -> Storage削除 -> Firestore削除を一気に行う
		if err := models.DeleteVocabWithAudio(ctx, firestoreClient, docId); err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, err)
			return
		}

	} else if mode == "all" {
		// --- 【一括削除（退会）】 ---
		// ① Storage内のユーザーフォルダを掃除
		// ※ models.DeleteAllAudioFiles(ctx, uid) を想定
		if err := models.DeleteAllAudioFile(ctx, uid); err != nil {
			fmt.Printf("Storage cleanup warning for user %s: %v\n", uid, err)
		}

		// ② Firestoreからユーザー取得
		user, err := models.GetUser(uid, ctx, firestoreClient)
		if err != nil {
			responses.SendErrorResponse(w, http.StatusNotFound, fmt.Errorf("user not found"))
			return
		}

		// ③ 全Vocabドキュメントを削除
		if err := models.DeleteAllUserVocabs(user.SelfRef, ctx, firestoreClient); err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to delete all vocabs: %v", err))
			return
		}

		// ④ ユーザードキュメント自体の削除
		if err := models.DeleteUserByUid(uid, ctx, firestoreClient); err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to delete user doc: %v", err))
			return
		}

		// ⑤ Firebase Authからユーザーを削除 (これを最後にすることで、途中でエラーが出ても再試行可能にする)
		if err := authClient.DeleteUser(ctx, uid); err != nil {
			responses.SendErrorResponse(w, http.StatusInternalServerError, fmt.Errorf("failed to delete auth user: %v", err))
			return
		}

	} else {
		responses.SendErrorResponse(w, http.StatusBadRequest, fmt.Errorf("invalid mode: %s", mode))
		return
	}

	// 6. レスポンス
	resMap := map[string]string{
		"message": "deletion successful",
		"mode":    mode,
	}
	resBody, _ := json.Marshal(resMap)
	responses.SendResponse(w, http.StatusOK, resBody)
}
