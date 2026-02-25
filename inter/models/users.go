package models

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
)

const UsersCollection = "users"

type Users struct {
	DocId          string                 `json:"uid"`
	SelfRef        *firestore.DocumentRef `json:"-"` // JSON出力には含めないため "-" を指定
	DisplayName    string                 `json:"display_name"`
	Email          string                 `json:"email"`
	PhotoUrl       string                 `json:"photo_url"`
	TargetLanguage string                 `json:"target_language"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

func GetUser(uid string, ctx context.Context, firestoreClient *firestore.Client) (*Users, error) {
	// 1. 指定したUIDのドキュメントを取得
	userSnapshot, err := firestoreClient.Collection(UsersCollection).Doc(uid).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	var user Users
	// 2. データを構造体にマッピング（ここではデータのみが入る）
	if err = userSnapshot.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to convert snapshot to users struct: %v", err)
	}

	// 3. リファレンスとIDをセット
	user.SelfRef = userSnapshot.Ref  // これでリファレンス（パス情報）が保持される
	user.DocId = userSnapshot.Ref.ID // 明示的にドキュメントIDをセット

	return &user, nil
}

func DeleteUserByUid(uid string, ctx context.Context, firestoreClient *firestore.Client) error {
	// 1. GetUser（先ほど作った関数）を利用して存在確認
	user, err := GetUser(uid, ctx, firestoreClient)
	if err != nil {
		// ドキュメントが存在しない場合もここでエラーになる
		return fmt.Errorf("could not find user before deletion: %v", err)
	}

	// 2. 取得したリファレンスを使って削除
	if _, err := user.SelfRef.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	return nil
}
