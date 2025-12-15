# Go API Starter

[![CI](https://github.com/zareh/go-api-starter/actions/workflows/ci.yml/badge.svg)](https://github.com/zareh/go-api-starter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/zareh/go-api-starter)](https://goreportcard.com/report/github.com/zareh/go-api-starter)

JWT認証付きTODO管理REST APIのスターターテンプレート。Goのベストプラクティスに基づいて設計されています。

## 🚀 機能

- **JWT認証**: セキュアなトークンベース認証
- **TODO CRUD**: 完全なCRUD操作
- **PostgreSQL**: 堅牢なデータベース
- **Docker対応**: ワンコマンドで起動可能
- **CI/CD**: GitHub Actionsによる自動テスト
- **クリーンアーキテクチャ**: Handler → Service → Repository の層構造

## 🛠️ 技術スタック

| カテゴリ | 技術 |
|----------|------|
| 言語 | Go 1.25+ |
| フレームワーク | Echo v4 |
| データベース | PostgreSQL 16 |
| 認証 | JWT (golang-jwt) |
| コンテナ | Docker, Docker Compose |
| CI/CD | GitHub Actions |

## 📁 ディレクトリ構造

```
go-api-starter/
├── cmd/server/          # エントリーポイント
├── internal/
│   ├── config/          # 設定管理
│   ├── handler/         # HTTPハンドラ
│   ├── model/           # データモデル
│   ├── repository/      # データベース操作
│   └── service/         # ビジネスロジック
├── db/
│   ├── migrations/      # DBマイグレーション
│   └── queries/         # sqlcクエリ
├── .github/workflows/   # CI設定
├── Dockerfile
├── docker-compose.yml
└── Makefile
```

## 🏃 クイックスタート

### 前提条件

- Go 1.22+
- Docker & Docker Compose
- Make

### セットアップ

```bash
# リポジトリをクローン
git clone https://github.com/zareh/go-api-starter.git
cd go-api-starter

# 環境変数を設定
cp .env.example .env

# 開発ツールをインストール
make setup

# Docker環境を起動
make docker-up

# マイグレーションを実行
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/go_api_starter?sslmode=disable"
make migrate

# アプリケーションを起動
make run
```

### Docker Composeで起動

```bash
docker-compose up -d
```

## 🔌 API仕様

### エンドポイント一覧

| メソッド | パス | 説明 | 認証 |
|----------|------|------|------|
| GET | `/health` | ヘルスチェック | 不要 |
| POST | `/api/v1/users` | ユーザー登録 | 不要 |
| POST | `/api/v1/auth/login` | ログイン | 不要 |
| GET | `/api/v1/todos` | TODO一覧取得 | **必要** |
| POST | `/api/v1/todos` | TODO作成 | **必要** |
| GET | `/api/v1/todos/:id` | TODO詳細取得 | **必要** |
| PUT | `/api/v1/todos/:id` | TODO更新 | **必要** |
| DELETE | `/api/v1/todos/:id` | TODO削除 | **必要** |

### 使用例

#### ユーザー登録

```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

#### ログイン

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "password": "password123"}'
```

#### TODO作成（認証必要）

```bash
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"title": "Learn Go"}'
```

## 🗄️ データベース設計

### ER図

```
┌──────────────────┐       ┌──────────────────┐
│      users       │       │      todos       │
├──────────────────┤       ├──────────────────┤
│ id (PK)          │───┐   │ id (PK)          │
│ email            │   │   │ user_id (FK)     │←─┘
│ password_hash    │   │   │ title            │
│ created_at       │   └──→│ completed        │
│ updated_at       │       │ created_at       │
└──────────────────┘       │ updated_at       │
                           └──────────────────┘
```

## 🧪 テスト

```bash
# 全テストを実行
make test

# カバレッジ付きで実行
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 開発コマンド

```bash
make build        # ビルド
make run          # 実行
make test         # テスト
make lint         # リンター実行
make docker-up    # Docker起動
make docker-down  # Docker停止
make migrate      # マイグレーション実行
make sqlc         # sqlcコード生成
```

## 📦 依存関係のインストール

```bash
go mod download
go mod tidy
```

## 🔧 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|------------|
| PORT | サーバーポート | 8080 |
| DATABASE_URL | PostgreSQL接続URL | - |
| JWT_SECRET | JWT署名キー | - |
| JWT_EXPIRY | JWTの有効期限（時間） | 24 |

## 📄 ライセンス

MIT License

## 🤝 コントリビューション

1. Fork する
2. Feature branch を作成する (`git checkout -b feature/amazing-feature`)
3. 変更をコミットする (`git commit -m 'Add amazing feature'`)
4. Branch を Push する (`git push origin feature/amazing-feature`)
5. Pull Request を作成する
