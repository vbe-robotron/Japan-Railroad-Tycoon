// Package store はゲーム状態の SQL 永続化を提供する。
//
// database/sql と modernc.org/sqlite を用いて、セーブデータや
// アクションログを読み書きする。純粋ロジックは core に置き、
// このパッケージは永続化の責務のみを負う。
package store
