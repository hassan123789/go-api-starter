# ADR-006: RBAC (Role-Based Access Control) Strategy

## Status
Accepted

## Context
APIのセキュリティを強化するため、ユーザーの役割に基づいたアクセス制御が必要です。リソースへのアクセスを適切に制限し、最小権限の原則を実装します。

## Decision
シンプルな3層ロールベースのアクセス制御を実装します。

### ロール定義

```go
type Role string

const (
    RoleAdmin  Role = "admin"   // 全権限
    RoleUser   Role = "user"    // 標準ユーザー
    RoleViewer Role = "viewer"  // 読み取り専用
)
```

### パーミッション定義

```go
type Permission string

const (
    // ユーザー関連
    PermUserRead   Permission = "user:read"
    PermUserCreate Permission = "user:create"
    PermUserUpdate Permission = "user:update"
    PermUserDelete Permission = "user:delete"

    // Todo関連
    PermTodoRead   Permission = "todo:read"
    PermTodoCreate Permission = "todo:create"
    PermTodoUpdate Permission = "todo:update"
    PermTodoDelete Permission = "todo:delete"

    // 管理者機能
    PermAdminAccess Permission = "admin:access"
)
```

### ロール-パーミッション マッピング

| Permission      | Admin | User | Viewer |
|-----------------|-------|------|--------|
| user:read       | ✅    | ✅   | ✅     |
| user:create     | ✅    | ❌   | ❌     |
| user:update     | ✅    | ⚠️*  | ❌     |
| user:delete     | ✅    | ❌   | ❌     |
| todo:read       | ✅    | ✅   | ✅     |
| todo:create     | ✅    | ✅   | ❌     |
| todo:update     | ✅    | ⚠️*  | ❌     |
| todo:delete     | ✅    | ⚠️*  | ❌     |
| admin:access    | ✅    | ❌   | ❌     |

*⚠️ = 自分のリソースのみ

### ミドルウェア使用例

```go
// 特定のパーミッションを要求
e.GET("/admin/users", handler, middleware.RequirePermission(cfg, rbac.PermAdminAccess))

// 最小ロールレベルを要求
e.POST("/todos", handler, middleware.RequireRole(cfg, rbac.RoleUser))

// リソースオーナーシップチェック
e.PUT("/todos/:id", handler, middleware.RequireResourceAccess(cfg, "todo", getOwnerID))
```

## Consequences

### Positive
- 明確な権限境界
- 最小権限の原則の実装
- 監査対応（誰が何にアクセスできるか明確）
- スケーラブルな設計

### Negative
- 実装の複雑さ増加
- パフォーマンスオーバーヘッド
- ロール追加時の更新作業

### Risks
- 権限エスカレーションの脆弱性
- ロール設定ミスによるアクセス漏洩

## References
- [OWASP Access Control Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Access_Control_Cheat_Sheet.html)
- [NIST RBAC Model](https://csrc.nist.gov/projects/role-based-access-control)
