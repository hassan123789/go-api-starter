# ADR-003: JWT認証戦略

## ステータス

✅ **Accepted** - 2025-12-29

## コンテキスト

REST APIの認証方式を選定する必要がありました。

検討した選択肢：

| 方式 | 説明 | ステートレス |
|------|------|------------|
| **Session Cookie** | サーバーサイドセッション | ❌ |
| **JWT (JSON Web Token)** | 自己完結型トークン | ✅ |
| **OAuth 2.0** | 外部認証プロバイダー連携 | ✅ |
| **API Key** | 静的キー認証 | ✅ |

## 決定

**JWT（JSON Web Token）** を採用し、以下の戦略を適用する：

### トークン構造

```json
{
  "header": {
    "alg": "HS256",
    "typ": "JWT"
  },
  "payload": {
    "user_id": 123,
    "email": "user@example.com",
    "exp": 1735500000,
    "iat": 1735413600
  },
  "signature": "..."
}
```

### 実装

```go
import (
    "github.com/golang-jwt/jwt/v5"
    echojwt "github.com/labstack/echo-jwt/v4"
)

// トークン生成
func GenerateToken(user *User, secret string, expiry time.Duration) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID,
        "email":   user.Email,
        "exp":     time.Now().Add(expiry).Unix(),
        "iat":     time.Now().Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// ミドルウェア設定
e.Use(echojwt.WithConfig(echojwt.Config{
    SigningKey:  []byte(config.JWTSecret),
    TokenLookup: "header:Authorization:Bearer ",
    ErrorHandler: func(c echo.Context, err error) error {
        return echo.NewHTTPError(http.StatusUnauthorized, "invalid or expired token")
    },
}))
```

## 理由

### 1. ステートレス設計

```
┌─────────┐                    ┌─────────┐
│ Client  │                    │ Server  │
└────┬────┘                    └────┬────┘
     │                              │
     │  POST /login                 │
     │  {email, password}           │
     │─────────────────────────────>│
     │                              │ Validate credentials
     │                              │ Generate JWT
     │  {token: "eyJ..."}           │
     │<─────────────────────────────│
     │                              │
     │  GET /todos                  │
     │  Authorization: Bearer eyJ...│
     │─────────────────────────────>│
     │                              │ Verify signature
     │                              │ No DB lookup needed!
     │  [{id: 1, title: "..."}]     │
     │<─────────────────────────────│
```

- サーバーにセッション状態を保持しない
- 水平スケーリングが容易
- Redis等のセッションストアが不要

### 2. 自己完結型

JWTにはユーザー情報が含まれるため、認証のたびにDBクエリが不要：

```go
// トークンからユーザーID取得（DB不要）
func getUserIDFromToken(c echo.Context) (int64, error) {
    token := c.Get("user").(*jwt.Token)
    claims := token.Claims.(jwt.MapClaims)
    return int64(claims["user_id"].(float64)), nil
}
```

### 3. クロスドメイン対応

- Cookie制約を回避
- モバイルアプリ、SPA、マイクロサービス間で統一的に使用可能

### 4. 署名による改ざん検知

```go
// HMAC-SHA256による署名
// ペイロードが改ざんされると署名検証が失敗
jwt.SigningMethodHS256
```

## セキュリティ設計

### 有効期限の設定

```go
const (
    AccessTokenExpiry  = 24 * time.Hour   // アクセストークン: 24時間
    RefreshTokenExpiry = 7 * 24 * time.Hour // リフレッシュトークン: 7日（将来実装）
)
```

### シークレットの管理

```go
// 環境変数から読み込み
type Config struct {
    JWTSecret string `env:"JWT_SECRET,required"`
}

// 本番環境では強力なシークレットを使用
// openssl rand -base64 64
```

### セキュリティヘッダー

```go
// レスポンスヘッダーでセキュリティ強化
e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
    XSSProtection:         "1; mode=block",
    ContentTypeNosniff:    "nosniff",
    XFrameOptions:         "DENY",
    HSTSMaxAge:            31536000,
    ContentSecurityPolicy: "default-src 'self'",
}))
```

## 将来の拡張計画

### Phase 2: リフレッシュトークン

```go
// アクセストークン + リフレッシュトークンの2トークン方式
type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int    `json:"expires_in"`
}

// POST /auth/refresh
func RefreshToken(c echo.Context) error {
    // リフレッシュトークン検証
    // 新しいアクセストークン発行
}
```

### Phase 3: トークン無効化（ブラックリスト）

```go
// Redis を使ったトークンブラックリスト
type TokenBlacklist interface {
    Add(tokenID string, expiry time.Duration) error
    IsBlacklisted(tokenID string) bool
}
```

## 結果

### ポジティブ

- ✅ スケーラブル（ステートレス）
- ✅ 認証時のDBクエリ不要
- ✅ モバイル・SPA・API間で統一的
- ✅ Echo公式のJWTミドルウェアと統合

### ネガティブ

- ⚠️ トークン無効化が即座にできない（有効期限まで有効）
- ⚠️ ペイロードサイズが大きくなると帯域を消費
- ⚠️ シークレット漏洩時のリスク

### 軽減策

- 短い有効期限（24時間）を設定
- 将来的にリフレッシュトークン + ブラックリストを実装
- シークレットは環境変数で管理、定期ローテーション

## 参考

- [JWT.io](https://jwt.io/)
- [RFC 7519 - JSON Web Token](https://datatracker.ietf.org/doc/html/rfc7519)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
