# ADR-002: Echo Frameworkの選定

## ステータス

✅ **Accepted** - 2025-12-29

## コンテキスト

Go REST APIを構築するためのHTTPフレームワークを選定する必要がありました。

検討した選択肢：

| フレームワーク | GitHub Stars | 特徴 |
|--------------|-------------|------|
| **net/http** | - | 標準ライブラリ、依存なし |
| **Gin** | 80k+ | 高速、最も人気 |
| **Echo** | 30k+ | 高速、ミドルウェア豊富 |
| **Chi** | 18k+ | 軽量、net/http互換 |
| **Fiber** | 35k+ | Express風、fasthttp使用 |

## 決定

**Echo Framework v4** を採用する。

```go
import "github.com/labstack/echo/v4"

e := echo.New()
e.Use(middleware.Logger())
e.Use(middleware.Recover())

e.GET("/health", healthHandler)
e.POST("/api/v1/todos", createTodoHandler)
```

## 理由

### 1. ミドルウェアエコシステム

Echoは公式で多くのミドルウェアを提供しており、すぐに使える：

```go
import (
    "github.com/labstack/echo/v4/middleware"
    echojwt "github.com/labstack/echo-jwt/v4"
)

// JWT認証 - 公式サポート
e.Use(echojwt.WithConfig(echojwt.Config{
    SigningKey: []byte(secret),
}))

// CORS
e.Use(middleware.CORS())

// リクエストID
e.Use(middleware.RequestID())

// レート制限
e.Use(middleware.RateLimiter(...))

// gzip圧縮
e.Use(middleware.Gzip())
```

### 2. パフォーマンス

Echoは高性能なルーティングを実現：

| フレームワーク | ベンチマーク (req/sec) |
|--------------|---------------------|
| Echo | 165,000+ |
| Gin | 160,000+ |
| Chi | 130,000+ |
| net/http | 100,000+ |

### 3. 優れたエラーハンドリング

```go
// カスタムHTTPエラーハンドラー
e.HTTPErrorHandler = func(err error, c echo.Context) {
    code := http.StatusInternalServerError
    message := "Internal Server Error"

    if he, ok := err.(*echo.HTTPError); ok {
        code = he.Code
        message = he.Message.(string)
    }

    c.JSON(code, map[string]string{"error": message})
}

// ハンドラー内でのエラー返却
func handler(c echo.Context) error {
    if err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "invalid input")
    }
    return c.JSON(http.StatusOK, result)
}
```

### 4. バリデーション統合

```go
import "github.com/go-playground/validator/v10"

type CustomValidator struct {
    validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
    return cv.validator.Struct(i)
}

e.Validator = &CustomValidator{validator: validator.New()}

// ハンドラー内
type CreateTodoRequest struct {
    Title string `json:"title" validate:"required,max=255"`
}

func createTodo(c echo.Context) error {
    req := new(CreateTodoRequest)
    if err := c.Bind(req); err != nil {
        return err
    }
    if err := c.Validate(req); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, err.Error())
    }
    // ...
}
```

### 5. コンテキスト拡張

```go
// JWT からユーザー情報取得
func getUserFromContext(c echo.Context) (*User, error) {
    token := c.Get("user").(*jwt.Token)
    claims := token.Claims.(jwt.MapClaims)
    userID := int64(claims["user_id"].(float64))
    return userID, nil
}
```

### 6. テスタビリティ

```go
import (
    "net/http/httptest"
    "github.com/labstack/echo/v4"
)

func TestCreateTodo(t *testing.T) {
    e := echo.New()
    req := httptest.NewRequest(http.MethodPost, "/todos", strings.NewReader(`{"title":"test"}`))
    req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
    rec := httptest.NewRecorder()
    c := e.NewContext(req, rec)

    if assert.NoError(t, handler.CreateTodo(c)) {
        assert.Equal(t, http.StatusCreated, rec.Code)
    }
}
```

## 却下した選択肢

### net/http（標準ライブラリ）

- ❌ ルーティングが弱い（パスパラメータなど）
- ❌ ミドルウェアを自前実装する必要
- ❌ JWT統合などの機能がない

### Gin

- ⚠️ 同等の機能を持つが、JWTミドルウェアがEchoの方が使いやすい
- ⚠️ Echoの方がエラーハンドリングが直感的

### Fiber

- ❌ `fasthttp`ベースで`net/http`と互換性がない
- ❌ 一部の標準ライブラリやミドルウェアが使えない

## 結果

### ポジティブ

- ✅ 公式JWTミドルウェアとの統合がスムーズ
- ✅ 豊富なミドルウェア（Rate Limit, CORS, Logger等）
- ✅ 高パフォーマンス
- ✅ 直感的なエラーハンドリング
- ✅ 活発なメンテナンス

### ネガティブ

- ⚠️ `net/http`との若干の違い（`http.Handler`ではなく`echo.HandlerFunc`）
- ⚠️ Ginほどの知名度・採用実績はない

## 参考

- [Echo Framework 公式ドキュメント](https://echo.labstack.com/)
- [Echo vs Gin Benchmark](https://github.com/smallnest/go-web-framework-benchmark)
- [echo-jwt GitHub](https://github.com/labstack/echo-jwt)
