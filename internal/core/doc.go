// Package core はゲームの純粋ロジックを提供する。
//
// このパッケージは HTTP・SQL・OS・time などに依存してはならない。
// 乱数はゲームの rng 経由のみを許可し、seed とアクション列から
// 同じ結果が再現される決定論を保証する。
package core
