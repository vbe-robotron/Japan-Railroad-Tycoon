package core

import "sort"

// このファイルは AI の基盤（構造体・実行順・実行エントリ）を提供する。
//
// Phase 8-1 では「史実ルート追従」のみを実装する。財務・M&A・開発 AI は
// Phase 8-2 以降で追加する。

// AI は 1 社の意思決定を担当する。
type AI struct {
	CompanyID   string      `json:"company_id"`
	Personality Personality `json:"personality"`
	HistoryMode HistoryMode `json:"history_mode"`
}

// HistoryMode は史実への追従度を表す。
type HistoryMode int

const (
	// HistoryStrict は史実通りに行動する。
	HistoryStrict HistoryMode = iota
	// HistoryLoose は史実を参考に行動する（Phase 8-2 以降）。
	HistoryLoose
	// HistoryIgnore は史実を無視して完全自由に行動する。
	HistoryIgnore
)

// RunAI は全社の AI を実行し、発生したイベントを返す。
//
// EndTurn から呼ばれる。プレイヤー会社および解散済みの会社はスキップする。
// 実行順は aiActionOrder により辞書順で安定化する（決定論のため）。
func (g *Game) RunAI() []Event {
	var events []Event
	order := g.aiActionOrder()
	for _, id := range order {
		if id == g.PlayerID {
			continue
		}
		c := g.Companies[id]
		if c == nil || c.Dissolved != nil {
			continue
		}
		ai := g.AIs[id]
		if ai == nil {
			continue
		}
		events = append(events, g.runAI(c, ai)...)
	}
	return events
}

// aiActionOrder は AI の実行順を返す（辞書順で安定）。
func (g *Game) aiActionOrder() []string {
	ids := make([]string, 0, len(g.Companies))
	for id := range g.Companies {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// runAI は 1 社の AI を実行する。
//
// 処理順序:
//  1. 史実ルート追従（HistoryIgnore 以外）
//  2. 適応ルート（Phase 8-2 以降、今回は何もしない）
func (g *Game) runAI(c *Company, ai *AI) []Event {
	// 史実ルート追従
	if ai.HistoryMode != HistoryIgnore {
		if evs := g.aiFollowHistory(c, ai); len(evs) > 0 {
			return evs
		}
	}

	// 適応ルート（Phase 8-2 以降）
	// 今回は何もしない
	return nil
}
