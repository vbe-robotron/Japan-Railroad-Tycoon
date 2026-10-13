// Package core は Japan Railroad Tycoon のゲームロジックを提供する。
//
// このパッケージは net/http, database/sql, time, os に依存しない。
// 純粋なゲームロジックのみを含み、永続化・HTTP・CLI は別パッケージが担う。
//
// 主な型:
//   - Game: ゲーム全体の状態
//   - Company: 鉄道会社
//   - Line: 路線
//   - City: 都市
//   - Event: ゲーム内イベント
//   - Action: プレイヤー or AI の行動
//
// ID設計:
//   - 静的データ（Node/City/Company/Event）: 意味的ID（"tokyo", "hankyu"）
//   - 動的データ（Line/Facility/Game/EventLog）: NanoID（21文字）
//
// 決定論:
//   - seed + アクション列 → 同じ結果
//   - 乱数は g.rng 経由のみ
package core
