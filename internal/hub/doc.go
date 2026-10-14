// Package hub はゲームセッションの管理と WebSocket 購読を提供する。
//
// 役割:
//   - 複数ゲームの同時管理
//   - ゲームごとのロック（Session.mu）
//   - 購読者へのイベント配信（Broadcast）
//
// 依存:
//   - core: ゲームロジック
//   - store: 永続化
package hub
