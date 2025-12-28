# ADR-004: エラーハンドリング設計

## ステータス

✅ **Accepted** - 2025-12-29

## コンテキスト

Go REST APIにおいて、一貫性のある堅牢なエラーハンドリング戦略を設計する必要がありました。

Goのエラーハンドリングの課題：

1. 標準の `error` インターフェースは情報が少ない
2. エラーの種類を判別しにくい
3. HTTPステータスコードとの紐付けが必要
4. スタックトレースがない

## 決定

**カスタムエラー型 + errors.Is/As** を使った型安全なエラーハンドリングを採用する。

### エラー定義

```go
// internal/model/errors.go

package model

import "errors"

// センチネルエラー（エラー種別の判定に使用）
var (
    ErrUserNotFound      = errors.New("user not found")
    ErrUserAlreadyExists = errors.New("user already exists")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrTodoNotFound      = errors.New("todo not found")
    ErrUnauthorized      = errors.New("unauthorized")
    ErrForbidden         = errors.New("forbidden")
    ErrValidation        = errors.New("validation error")
)

// AppError - アプリケーションエラー型
type AppError struct {
    Code    string // エラーコード（"USER_NOT_FOUND"等）
    Message string // ユーザー向けメッセージ
    Err     error  // 元のエラー（ラップ）
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Message + ": " + e.Err.Error()
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// エラー生成ヘルパー
func NewAppError(code, message string, err error) *AppError {
    return &AppError{
        Code:    code,
        Message: message,
        Err:     err,
    }
}
```

### エラー判定パターン

```go
// Service層でのエラー生成
func (s *AuthService) Login(ctx context.Context, email, password string) (*User, error) {
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, ErrInvalidCredentials
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    if !checkPassword(user.Password, password) {
        return nil, ErrInvalidCredentials
    }

    return user, nil
}

// Handler層でのエラー判定
func (h *AuthHandler) Login(c echo.Context) error {
    // ...
    user, err := h.authService.Login(ctx, req.Email, req.Password)
    if err != nil {
        switch {
        case errors.Is(err, model.ErrInvalidCredentials):
            return echo.NewHTTPError(http.StatusUnauthorized, "invalid email or password")
        case errors.Is(err, model.ErrUserNotFound):
            return echo.NewHTTPError(http.StatusNotFound, "user not found")
        default:
            h.logger.Error("login failed", "error", err)
            return echo.NewHTTPError(http.StatusInternalServerError, "internal server error")
        }
    }
    // ...
}
```

## 理由

### 1. Go 1.13+ の errors.Is/As サポート

```go
// errors.Is - エラーの種類を判定
if errors.Is(err, ErrUserNotFound) {
    // ユーザーが見つからない場合の処理
}

// errors.As - エラーを特定の型に変換
var appErr *AppError
if errors.As(err, &appErr) {
    fmt.Println(appErr.Code)    // "USER_NOT_FOUND"
    fmt.Println(appErr.Message) // "ユーザーが見つかりません"
}
```

### 2. エラーのラップ（コンテキスト追加）

```go
// fmt.Errorf の %w でエラーをラップ
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*User, error) {
    user, err := r.db.QueryRow(ctx, query, id)
    if err != nil {
        return nil, fmt.Errorf("UserRepository.GetByID(%d): %w", id, err)
    }
    return user, nil
}

// ログ出力
// "UserRepository.GetByID(123): sql: no rows in result set"
```

### 3. HTTPステータスコードのマッピング

```go
// エラー → HTTPステータスコードの変換
func errorToHTTPStatus(err error) int {
    switch {
    case errors.Is(err, model.ErrUserNotFound),
         errors.Is(err, model.ErrTodoNotFound):
        return http.StatusNotFound
    case errors.Is(err, model.ErrInvalidCredentials):
        return http.StatusUnauthorized
    case errors.Is(err, model.ErrUserAlreadyExists):
        return http.StatusConflict
    case errors.Is(err, model.ErrValidation):
        return http.StatusBadRequest
    case errors.Is(err, model.ErrForbidden):
        return http.StatusForbidden
    default:
        return http.StatusInternalServerError
    }
}
```

### 4. グローバルエラーハンドラー

```go
// Echo のカスタムエラーハンドラー
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := http.StatusInternalServerError
    message := "Internal Server Error"

    // Echo HTTPError
    if he, ok := err.(*echo.HTTPError); ok {
        code = he.Code
        if m, ok := he.Message.(string); ok {
            message = m
        }
    }

    // AppError
    var appErr *model.AppError
    if errors.As(err, &appErr) {
        code = errorToHTTPStatus(appErr.Err)
        message = appErr.Message
    }

    // JSON レスポンス
    if !c.Response().Committed {
        c.JSON(code, map[string]interface{}{
            "error": message,
            "code":  code,
        })
    }
}
```

## エラーレスポンス形式

### 標準レスポンス

```json
{
    "error": "invalid email or password"
}
```

### 詳細レスポンス（開発環境のみ）

```json
{
    "error": "validation failed",
    "code": "VALIDATION_ERROR",
    "details": [
        {"field": "email", "message": "invalid email format"},
        {"field": "password", "message": "must be at least 8 characters"}
    ]
}
```

## 層別の責任

| 層 | 責任 |
|---|------|
| **Repository** | SQLエラーをラップ、センチネルエラーに変換 |
| **Service** | ビジネスルール違反をセンチネルエラーで返す |
| **Handler** | errors.Isで判定し、HTTPステータスに変換 |

```
Repository:  sql.ErrNoRows → model.ErrUserNotFound
Service:     model.ErrUserNotFound → そのまま返す
Handler:     model.ErrUserNotFound → 404 Not Found
```

## 結果

### ポジティブ

- ✅ 型安全なエラー判定（`errors.Is`）
- ✅ エラーの文脈が保持される（ラップ）
- ✅ 一貫したHTTPレスポンス形式
- ✅ ログに詳細なスタックトレース

### ネガティブ

- ⚠️ センチネルエラーの定義が増える
- ⚠️ 各層でエラーハンドリングコードが必要

### 軽減策

- エラー変換のヘルパー関数を用意
- グローバルエラーハンドラーで共通処理

## テストでのエラー検証

```go
func TestLogin_InvalidCredentials(t *testing.T) {
    // Arrange
    mockRepo := &MockUserRepository{
        GetByEmailFunc: func(ctx context.Context, email string) (*User, error) {
            return nil, model.ErrUserNotFound
        },
    }
    service := NewAuthService(mockRepo)

    // Act
    _, err := service.Login(context.Background(), "test@example.com", "wrong")

    // Assert
    assert.Error(t, err)
    assert.True(t, errors.Is(err, model.ErrInvalidCredentials))
}
```

## 参考

- [Go Blog - Working with Errors in Go 1.13](https://go.dev/blog/go1.13-errors)
- [Error Handling in Go](https://go.dev/doc/effective_go#errors)
- [Dave Cheney - Don't just check errors, handle them gracefully](https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully)
