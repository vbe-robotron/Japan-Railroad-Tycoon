# 進捗

## 完了
- [x] Phase 0-1: プロジェクト初期化
- [x] Phase 0-2: 共通型
- [x] Phase 1-1: 基本構造体
- [x] Phase 1-2: Event型
- [x] Phase 1-3: Action インターフェース
- [x] Phase 1-4: EndTurnAction
- [x] Phase 1-5: Sim ヘルパー
- [x] Phase 1-6: BuildLineAction
- [x] Phase 1-7: 経済シミュレーション（旅客収入・維持費・減価償却・赤字検出）

## 進行中
- [ ] Phase 1-8: 残りアクション

## 次
- [ ] Phase 1-8: 残りアクション（施設・M&A・廃線・貨物/不動産/流通収入）
- [ ] Phase 1-7.5: 経済バランス調整（必要なら）

## メモ
- module: github.com/you/jrt
- Go 1.22
- core は誰にも依存しない（禁止 import: net/http, database/sql, time, os）
- IDGenerator は決定論的 NanoID（math/rand + FNV-1a）。go-nanoid の Generate は
  custom reader 非対応のため core では外部依存を使わず自前生成している
- Phase 1-7 の経済:
  - 旅客収入のみ実装（貨物・不動産・流通は Phase 1-8）
  - 需要（Game.Demand）は日次需要として扱い、月次旅客数へ DaysPerMonth(=30) を
    掛けて換算（輸送力上限も月次換算）。統合シナリオを通すための係数調整。
  - EndTurn は「時間進行前の Date」で経済計算 → turn_ended → 時間進行 →
    year_changed の順
  - simulateEconomy は会社 ID 昇順で処理（map の順序不定を回避し決定論を担保）
  - City に Lat/Lon（実座標）を追加。cityLatLon は未設定時のみ従来のハッシュに
    フォールバック（既存テスト互換）
  - 現状は黒字路線が過大に儲かる（営業係数 4.5 等）。維持費・運賃・需要の
    バランスは Phase 1-7.5 で調整する
