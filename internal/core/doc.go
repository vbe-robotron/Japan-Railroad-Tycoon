// Package core は Japan Railroad Tycoon のゲームロジックを提供する。
//
// このパッケージは net/http, database/sql, time, os に依存しない。
// 純粋なゲームロジックのみを含み、永続化・HTTP・CLI は別パッケージが担う。
//
// 主な型:
//   - Game: ゲーム全体の状態
//   - Company: 鉄道会社
//   - Line: 路線
//   - City: 都市（市区町村代表駅）
//   - Facility: 開発施設（百貨店・住宅・遊園地など）
//   - Event: ゲーム内イベント
//   - Action: プレイヤー or AI の行動
//
// イベント:
//   - Event: ゲーム内イベント（全アクションが返す）
//   - 60種以上のタイプ（Event* 定数）
//   - 主要20種は Payload を構造体化
//   - ビルダー: With, ForCompany, ForLine, ForCity
//
// アクション:
//   - Action インターフェース（Type, Validate, Apply）
//   - 25種のアクション（建設・運用・財務・M&A・開発・政治・廃線・ターン）
//   - DecodeAction で type 文字列から復元
//   - Validate は状態を変更しない
//   - Apply は []Event を返す
//
// ターン進行:
//   - Game.EndTurn(): 1ヶ月進める
//   - EndTurnAction: プレイヤーの end-turn コマンド
//   - turn_ended イベント: ターン終了時に生成
//   - year_changed イベント: 年越し時に生成
//
// ID設計:
//   - 静的データ（Node/City/Company/Event）: 意味的ID（"tokyo", "hankyu"）
//   - 動的データ（Line/Facility/Game/EventLog）: NanoID（21文字）
//
// 決定論:
//   - seed + アクション列 → 同じ結果
//   - 乱数は g.rng 経由のみ
//
// シミュレーション:
//   - Game.DistanceKm: Haversine 距離 × 地形係数
//   - Game.LineCost: 敷設コスト
//   - Game.DefaultFare: デフォルト運賃
//   - Game.Demand: 旅客需要
//   - Game.FacilityCost: 施設建設コスト
//   - Terrain: 地形パラメータ
package core
