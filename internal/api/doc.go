// Package api は HTTP サーバーとクライアントを提供する。
//
// サーバー:
//   - NewRouter(h, st): HTTP ハンドラ
//   - Response 型: 統一レスポンス形式
//
// クライアント:
//   - NewClient(cfg): HTTP クライアント
//   - Client.ForGame(id): ゲーム単位のクライアント
//   - 全アクション・状態取得をメソッド化
package api
