package controller

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/models"
	"github.com/Yoshiki-programming/vokanote-backend.git/inter/utils/Slog"
	"net/http"
)

func DeleteUserController(ctx context.Context, firestoreClient *firestore.Client, uid string) (int, []byte, error) {
	Slog.DebugInfo("Start DeleteUserController for UID", uid)

	// 1. ユーザーの存在確認と情報の取得
	user, err := models.GetUser(uid, ctx, firestoreClient)
	if err != nil {
		Slog.Error(fmt.Errorf("failed to find user for deletion: %v", err))
		// ユーザーが見つからない場合は 404
		return http.StatusNotFound, nil, err
	}

	// --- 削除フェーズ ---
	// 2. 関連する Recommendations（おすすめ曲）を全削除
	if err := models.DeleteAllUserRecommendations(user.SelfRef, ctx, firestoreClient); err != nil {
		Slog.Error(fmt.Errorf("failed to delete recommendations: %v", err))
		return http.StatusInternalServerError, nil, err
	}
	Slog.DebugInfo("DELETED", "Recommendations for user: "+uid)

	// 3. 関連する Vocabs（単語帳）を全削除
	if err := models.DeleteAllUserVocabs(user.SelfRef, ctx, firestoreClient); err != nil {
		Slog.Error(fmt.Errorf("failed to delete vocabs: %v", err))
		return http.StatusInternalServerError, nil, err
	}
	Slog.DebugInfo("DELETED", "Vocabs for user: "+uid)

	// 4. 最後に Users ドキュメント本体を削除
	if err := models.DeleteUserByUid(user.SelfRef.ID, ctx, firestoreClient); err != nil {
		Slog.Error(fmt.Errorf("failed to delete user document: %v", err))
		return http.StatusInternalServerError, nil, err
	}
	Slog.DebugInfo("DELETED", "User document: "+uid)

	// すべて成功
	responseBody := []byte(`{"message": "user and all related data deleted successfully"}`)
	return http.StatusOK, responseBody, nil
}
