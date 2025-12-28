# ADR-007: Resilience Patterns (Circuit Breaker, Retry, Rate Limiting)

## Status
Accepted

## Context
外部依存サービス（データベース、外部API）の障害時にシステム全体が影響を受けることを防ぐため、レジリエンスパターンの実装が必要です。

## Decision
Circuit Breaker、Retry、Rate Limitingの3つのパターンを実装し、graceful degradationを実現します。

### 1. Circuit Breaker（遮断器パターン）

障害が連続した場合にリクエストを遮断し、システムを保護します。

```go
cb := resilience.NewCircuitBreaker("database",
    resilience.WithMaxFailures(5),      // 5回失敗で開く
    resilience.WithTimeout(30*time.Second), // 30秒後に半開
    resilience.WithFallback(func(ctx context.Context, err error) (interface{}, error) {
        return cachedData, nil // キャッシュからフォールバック
    }),
)

result, err := cb.Execute(ctx, func(ctx context.Context) (interface{}, error) {
    return db.Query(ctx, query)
})
```

**状態遷移**:
```
Closed → (failures >= max) → Open → (timeout) → Half-Open → (success) → Closed
                                                           → (failure) → Open
```

### 2. Retry with Backoff（リトライパターン）

一時的な障害に対して、指数バックオフでリトライします。

```go
retryer := resilience.NewRetryer(
    resilience.WithMaxAttempts(3),
    resilience.WithBackoff(&resilience.ExponentialBackoff{
        Initial:    100 * time.Millisecond,
        Max:        10 * time.Second,
        Multiplier: 2.0,
        Jitter:     true, // 同時リトライ回避
    }),
)

err := retryer.Do(ctx, func(ctx context.Context) error {
    return externalAPI.Call(ctx)
})
```

### 3. Rate Limiting（レート制限）

リクエストレートを制限してサービスを保護します。

```go
// グローバルレートリミッター
limiter := resilience.NewRateLimiter(100, 10) // 100 req/sec, バースト10

// キー別（IPアドレス別など）
keyedLimiter := resilience.NewKeyedRateLimiter(10, 5, time.Minute)
if !keyedLimiter.Allow(clientIP) {
    return ErrRateLimited
}
```

### パターンの組み合わせ

```go
// 推奨: Rate Limit → Circuit Breaker → Retry
func callExternalService(ctx context.Context) error {
    // 1. レート制限チェック
    if !rateLimiter.Allow() {
        return ErrRateLimited
    }

    // 2. サーキットブレーカー経由
    _, err := circuitBreaker.Execute(ctx, func(ctx context.Context) (interface{}, error) {
        // 3. リトライ付きで呼び出し
        return nil, retryer.Do(ctx, func(ctx context.Context) error {
            return externalAPI.Call(ctx)
        })
    })
    return err
}
```

## Consequences

### Positive
- カスケード障害の防止
- リソースの保護
- graceful degradation
- 自己回復能力

### Negative
- 実装の複雑さ
- パラメータチューニングの必要性
- フォールバック戦略の設計が必要

### Risks
- 過度に積極的なサーキットブレーカー設定
- リトライストームによる負荷増大

## References
- [Microsoft Azure - Circuit Breaker Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker)
- [Martin Fowler - CircuitBreaker](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Token Bucket Algorithm](https://en.wikipedia.org/wiki/Token_bucket)
