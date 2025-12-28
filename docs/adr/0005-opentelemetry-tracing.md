# ADR-005: OpenTelemetry Tracing Strategy

## Status
Accepted

## Context
分散システムでは、リクエストが複数のサービスを横断するため、パフォーマンスのボトルネックや障害の特定が困難です。効果的なオブザーバビリティを確保するため、分散トレーシングの導入が必要です。

## Decision
OpenTelemetryをトレーシングフレームワークとして採用し、Jaegerをバックエンドとして使用します。

### 選択理由

1. **OpenTelemetry（OTel）**
   - CNCFプロジェクト、ベンダーニュートラル
   - トレース・メトリクス・ログの統一API
   - 豊富なSDKとエコシステム
   - W3C Trace Context標準準拠

2. **Jaeger**
   - 高機能なUI
   - Dockerで容易にローカル実行
   - OTLP直接サポート
   - 本番環境でのスケーラビリティ

### 実装パターン

```go
// プロバイダの初期化
provider, err := tracing.NewProvider(&tracing.Config{
    ServiceName: "go-api-starter",
    Endpoint:    "localhost:4318",
    SampleRate:  1.0, // 開発環境では全てサンプリング
})
defer provider.Shutdown(ctx)

// Echoミドルウェア
e.Use(tracing.Middleware("go-api-starter"))

// DBトレーシング
tracer := tracing.NewDBTracer("go-api-starter", "postgres")
ctx, end := tracer.TraceQuery(ctx, "SELECT", "users", query)
defer end(err)
```

### サンプリング戦略
- **開発環境**: 100%（全リクエスト）
- **ステージング**: 10%
- **本番環境**: 1-5%（負荷に応じて調整）

### トレースに含める情報
- HTTP: メソッド、URL、ステータスコード、レイテンシ
- DB: 操作種別、テーブル名、影響行数
- ユーザー: ID、ロール（センシティブ情報は除外）
- エラー: スタックトレース、エラーメッセージ

## Consequences

### Positive
- リクエストフローの可視化
- パフォーマンスボトルネックの特定
- 障害の迅速な原因特定
- マイクロサービス移行への準備

### Negative
- オーバーヘッドの増加（1-5%）
- 追加のインフラストラクチャ要件
- データストレージコスト

### Risks
- サンプリングレートが高すぎるとパフォーマンス影響
- センシティブ情報のログ漏洩

## References
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [W3C Trace Context](https://www.w3.org/TR/trace-context/)
