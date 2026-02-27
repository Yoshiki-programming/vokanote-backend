# 修正前: FROM golang:1.22.5 as builder
FROM golang:1.25-alpine as builder
# ビルドに必要なツールをインストール
RUN apk add --no-cache git

# 作業ディレクトリを設定
# パスは任意ですが、プロジェクト名に合わせると管理しやすいです
WORKDIR /app

# 先に依存関係をコピーしてキャッシュを効かせる（ビルド高速化）
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコンテナにコピー
COPY . .

# ビルド実行（CGOを無効にしてスタティックバイナリを作成）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main cmd/main.go

# 実行環境
FROM alpine:latest
RUN apk add --no-cache ca-certificates

# ビルドしたバイナリをコピー
COPY --from=builder /main /main

# 設定ファイルと鍵ファイルをコピー
# プロジェクト名が vokanote-backend の場合、コピー元パスに注意
COPY --from=builder /app/configs /configs
COPY --from=builder /app/keys /keys

# 環境変数のデフォルト値（必要に応じて）
ENV GO_ENV=local
ENV PORT=8080

# ポートを公開
EXPOSE 8080

# 実行コマンド
CMD ["/main"]
