# ADR-001: Clean Architectureの採用

## ステータス

✅ **Accepted** - 2025-12-29

## コンテキスト

Go REST APIプロジェクトにおいて、長期的なメンテナビリティとテスタビリティを確保するためのアーキテクチャパターンを選定する必要がありました。

検討した選択肢：

1. **フラット構造** - すべてを1つのパッケージに配置
2. **MVC** - Model-View-Controller
3. **Clean Architecture** - Handler → Service → Repository の3層構造
4. **Hexagonal Architecture (Ports & Adapters)** - ポートとアダプター

## 決定

**Clean Architecture（3層アーキテクチャ）** を採用する。

```
┌─────────────────────────────────────────────────────────┐
│                    Handler Layer                         │
│  - HTTPリクエストのパース                                 │
│  - レスポンスのフォーマット                               │
│  - バリデーション                                         │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                    Service Layer                         │
│  - ビジネスロジック                                       │
│  - トランザクション管理                                   │
│  - 外部サービス連携                                       │
└─────────────────────────────────────────────────────────┘
                          ↓
┌─────────────────────────────────────────────────────────┐
│                   Repository Layer                       │
│  - データベースアクセス                                   │
│  - SQLクエリの実行                                        │
│  - エンティティのマッピング                               │
└─────────────────────────────────────────────────────────┘
```

## 理由

### 1. テスタビリティ

- 各層がインターフェースで分離されているため、モックを使った単体テストが容易
- Repositoryをモックすれば、データベースなしでService層をテスト可能

```go
// インターフェースによる分離
type TodoRepository interface {
    Create(ctx context.Context, todo *Todo) error
    GetByID(ctx context.Context, id int64) (*Todo, error)
    // ...
}

// テスト時はモックに差し替え
type MockTodoRepository struct {
    todos map[int64]*Todo
}
```

### 2. 関心の分離

- 各層が単一の責任を持つ
- HTTPの詳細はHandler層に閉じ込められる
- ビジネスルールの変更はService層のみ

### 3. 依存性の方向

- 依存は常に外側から内側へ
- 内側の層（Service、Repository）は外側の層（Handler）を知らない
- データベース変更の影響がService層に波及しない

### 4. Goエコシステムとの親和性

- 標準的なGoプロジェクトレイアウトと整合
- `internal/` ディレクトリでカプセル化
- インターフェースベースのDIはGoのイディオム

## 結果

### ポジティブ

- ✅ モックを使った高速な単体テストが可能
- ✅ 新メンバーがコードを理解しやすい（役割が明確）
- ✅ ビジネスロジックの再利用が容易
- ✅ データベースの変更（PostgreSQL → MySQL等）が比較的容易

### ネガティブ

- ⚠️ 小規模プロジェクトでは過剰設計に見える可能性
- ⚠️ 層を跨ぐ際のデータ変換（DTO ↔ Entity）が必要
- ⚠️ ボイラープレートコードが増える

### 軽減策

- Genericsを活用してボイラープレートを削減
- 小さな機能はシンプルに保つ（過度な抽象化を避ける）

## ディレクトリ構造

```
internal/
├── handler/        # Handler層 - HTTP関連
│   ├── auth.go
│   ├── todo.go
│   └── health.go
├── service/        # Service層 - ビジネスロジック
│   ├── auth.go
│   └── todo.go
├── repository/     # Repository層 - データアクセス
│   ├── user.go
│   └── todo.go
├── model/          # ドメインモデル
│   ├── user.go
│   ├── todo.go
│   └── errors.go
├── middleware/     # ミドルウェア
│   ├── auth.go
│   └── logging.go
└── config/         # 設定
    └── config.go
```

## 参考

- [The Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Project Structure Best Practices](https://go.dev/doc/modules/layout)
