# Architecture Decision Records (ADR)

このディレクトリには、プロジェクトのアーキテクチャに関する重要な決定事項を記録しています。

## ADRとは？

Architecture Decision Records（ADR）は、ソフトウェアアーキテクチャに関する重要な決定と、その背景・理由を記録するドキュメントです。

## 一覧

| ADR | タイトル | ステータス | 日付 |
|-----|---------|-----------|------|
| [001](./001-use-clean-architecture.md) | Clean Architectureの採用 | ✅ Accepted | 2025-12-29 |
| [002](./002-choose-echo-framework.md) | Echo Frameworkの選定 | ✅ Accepted | 2025-12-29 |
| [003](./003-jwt-authentication-strategy.md) | JWT認証戦略 | ✅ Accepted | 2025-12-29 |
| [004](./004-error-handling-approach.md) | エラーハンドリング設計 | ✅ Accepted | 2025-12-29 |

## ADRフォーマット

新しいADRを追加する場合は、以下のテンプレートを使用してください：

```markdown
# ADR-XXX: タイトル

## ステータス
Proposed | Accepted | Deprecated | Superseded

## コンテキスト
決定が必要になった背景・状況

## 決定
採用した決定内容

## 理由
その決定を選んだ理由

## 結果
決定による影響（ポジティブ・ネガティブ両方）

## 参考
関連するドキュメントやリンク
```

## 参考資料

- [ADR GitHub Organization](https://adr.github.io/)
- [Documenting Architecture Decisions - Michael Nygard](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
