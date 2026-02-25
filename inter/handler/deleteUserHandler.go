package handler

import (
	"context"
	"fmt"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/controller"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/utils/Slog"
	"google.golang.org/api/option"
	"net/http"
	"os"
	"strings"

	"github.com/Yoshiki-programming/vokanote-backend.git/inter/models"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/responses"
)

var Opt = option.WithCredentialsFile(os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	// CORS ヘッダーの設定
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// POST メソッド以外を弾く
	if r.Method != http.MethodPost {
		responses.SendErrorResponse(w, http.StatusMethodNotAllowed, fmt.Errorf("%s method not allowed", r.Method))
		return
	}

	ctx := context.Background()

	// 1. 各種クライアントの初期化
	firestoreClient, err := models.GetFirestoreClient(ctx, Opt, os.Getenv("PROJECT_ID"))
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer firestoreClient.Close() // クライアントのクローズを忘れずに

	authClient, err := models.GetAuthClient(ctx, Opt)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	// 2. Authorization ヘッダーから ID トークンを取得・検証
	authHeader := r.Header.Get("Authorization")
	idToken := strings.Replace(authHeader, "Bearer ", "", 1)
	token, err := authClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		responses.SendErrorResponse(w, http.StatusUnauthorized, fmt.Errorf("invalid token: %v", err))
		return
	}
	uid := token.UID

	// 3. Controller を呼び出して Firestore 側の関連データを一括削除
	// (先ほど作成した DeleteUserController を呼び出す)
	statusCode, body, err := controller.DeleteUserController(ctx, firestoreClient, uid)
	if err != nil {
		responses.SendErrorResponse(w, statusCode, err)
		return
	}

	// 4. Firestore の削除に成功したら、最後に Firebase Auth 側のユーザーを削除
	// これにより、二度と同じトークンでログインできなくなります
	if err := authClient.DeleteUser(ctx, uid); err != nil {
		// Firestore 削除後なので、ここでのエラーはログに留めるか、慎重に扱う
		Slog.Error(fmt.Errorf("failed to delete auth user: %v", err))
		responses.SendErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	Slog.DebugInfo("COMPLETED", "User account and all data fully purged for UID: "+uid)

	// 5. 成功レスポンスを送信
	responses.SendResponse(w, statusCode, body)
}
