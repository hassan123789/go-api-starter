# ADR-008: Audit Logging Strategy

## Status
Accepted

## Context
セキュリティ監査、コンプライアンス、問題調査のために、ユーザーアクションの追跡可能性が必要です。特に認証イベント、リソースの変更、管理者アクションを記録します。

## Decision
構造化された監査ログシステムを実装し、slogと永続ストレージの両方に記録します。

### 監査イベント構造

```go
type Event struct {
    ID           int64                  `json:"id"`
    Timestamp    time.Time              `json:"timestamp"`
    Action       Action                 `json:"action"`
    Severity     Severity               `json:"severity"`
    UserID       *int64                 `json:"user_id,omitempty"`
    UserEmail    string                 `json:"user_email,omitempty"`
    ResourceType string                 `json:"resource_type,omitempty"`
    ResourceID   *int64                 `json:"resource_id,omitempty"`
    Details      map[string]interface{} `json:"details,omitempty"`
    IPAddress    net.IP                 `json:"ip_address,omitempty"`
    UserAgent    string                 `json:"user_agent,omitempty"`
    RequestID    string                 `json:"request_id,omitempty"`
    Success      bool                   `json:"success"`
    Error        string                 `json:"error,omitempty"`
}
```

### アクション分類

| Category | Actions |
|----------|---------|
| 認証 | login, login_failed, logout, token_refresh, password_reset |
| ユーザー | user.create, user.update, user.delete |
| Todo | todo.create, todo.update, todo.delete |
| 管理者 | admin.access, role_change, config_change, audit_export |

### 深刻度レベル

```go
const (
    SeverityInfo     Severity = "info"     // 通常操作
    SeverityWarning  Severity = "warning"  // 注意が必要
    SeverityError    Severity = "error"    // エラー発生
    SeverityCritical Severity = "critical" // セキュリティ違反
)
```

### 使用例

```go
// ログイン成功
audit.Log(ctx, audit.LoginSuccess(userID, email, ip, userAgent, requestID))

// ログイン失敗
audit.Log(ctx, audit.LoginFailed(email, ip, userAgent, requestID, err))

// リソース作成
audit.Log(ctx, audit.ResourceCreated(audit.ActionTodoCreate, userID, email, "todo", todoID, requestID))

// Fluent Builder
event := audit.NewEvent(audit.ActionRoleChange).
    WithUser(userID, email).
    WithDetails(map[string]interface{}{
        "old_role": "user",
        "new_role": "admin",
    }).
    WithSeverity(audit.SeverityCritical).
    Build()
audit.Log(ctx, event)
```

### 記録すべきイベント

1. **必須（MUST）**
   - すべての認証イベント
   - ロール/権限の変更
   - 管理者アクション
   - セキュリティ関連の設定変更

2. **推奨（SHOULD）**
   - リソースの作成/更新/削除
   - APIエラー
   - レート制限発動

3. **オプション（MAY）**
   - 読み取り操作
   - 正常なAPI呼び出し

### 保持ポリシー

| Severity | 保持期間 |
|----------|---------|
| Critical | 7年 |
| Error/Warning | 1年 |
| Info | 90日 |

## Consequences

### Positive
- セキュリティインシデント調査
- コンプライアンス対応
- ユーザー行動分析
- 問題の迅速な特定

### Negative
- ストレージコスト
- パフォーマンスオーバーヘッド
- PII管理の複雑さ

### Risks
- ログにセンシティブ情報が含まれる
- ログの改ざん
- ストレージの枯渇

## References
- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
- [NIST SP 800-92 Guide to Computer Security Log Management](https://csrc.nist.gov/publications/detail/sp/800-92/final)
