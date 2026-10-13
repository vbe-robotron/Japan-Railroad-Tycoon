package core

import "fmt"

// EndTurn は1ターン分の時間を進め、発生したイベントを返す。
//
// 処理順序:
//  1. 経済シミュレーション（時間進行前の Date で計算）
//  2. turn_ended（時間進行前の Date/Turn）
//  3. 時間進行（Date +1ヶ月、Turn +1）
//  4. year_changed（時間進行後の Date/Turn）
//
// 決定論: seed + アクション列から常に同じイベント列が生成される。
func (g *Game) EndTurn() []Event {
	var events []Event

	// 1. 経済シミュレーション
	//
	// 収益・維持費・減価償却は「このターン（進行前の月）」のものとして計算する。
	events = append(events, g.simulateEconomy()...)

	// 2. 現時点の日付で turn_ended イベントを生成
	//    （「このターンが終わった」という意味なので、時間進行前の Date を使う）
	events = append(events, g.Info(
		EventTurnEnded,
		fmt.Sprintf("%s ターン終了", g.Date.String()),
	).With("turn", g.Turn))

	// 3. 年越し判定用に前の年を記録
	prevYear := g.Date.Year

	// 4. 時間を1ヶ月進める
	g.Date = g.Date.AddMonths(1)
	g.Turn++

	// 5. 年越し判定
	//
	// 設計判断:
	//   - turn_ended の Date/Turn は時間進行前の値
	//     （「1910年3月が終わった」という意味）
	//   - year_changed の Date/Turn は時間進行後の値
	//     （「1911年になった」という意味）
	if g.Date.Year != prevYear {
		events = append(events, g.Info(
			EventYearChanged,
			fmt.Sprintf("%d年になりました", g.Date.Year),
		).With("year", g.Date.Year))
	}

	// 6. イベントを記録
	g.AppendEvents(events...)

	return events
}
