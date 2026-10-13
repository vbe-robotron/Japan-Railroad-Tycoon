# 進捗

## 完了
- [x] Phase 0-1: プロジェクト初期化
- [x] Phase 0-2: 共通型

## 進行中
- [ ] Phase 1-1: 基本構造体

## 次
- [ ] Phase 1-1: Game/Company/Line/City

## メモ
- module: github.com/you/jrt
- Go 1.22
- core は誰にも依存しない
- IDGenerator は決定論的 NanoID（math/rand + FNV-1a）。go-nanoid の Generate は
  custom reader 非対応のため core では外部依存を使わず自前生成している
