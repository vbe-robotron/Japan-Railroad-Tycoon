package cli

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"

	"github.com/you/jrt/internal/api"
	"github.com/you/jrt/internal/core"
)

// Stdout / Stderr は出力先。テストから差し替え可能にする。
var (
	Stdout io.Writer = os.Stdout
	Stderr io.Writer = os.Stderr

	// outputMu は WebSocket イベントとコマンド出力の混線を防ぐ。
	outputMu sync.Mutex
)

var (
	colorSuccess = color.New(color.FgGreen)
	colorWarn    = color.New(color.FgYellow)
	colorDanger  = color.New(color.FgRed)
	colorBold    = color.New(color.Bold)
	colorInfo    = color.New(color.FgCyan)
)

// formatYen は金額を日本語単位付きの文字列へ整形する。
func formatYen(v int64) string { return core.FormatYen(v) }

// printSection は見出しを太字で書き出す。
func printSection(w io.Writer, title string) {
	colorBold.Fprintln(w, title)
}

// printKV は "ラベル: 値" 形式で書き出す。
func printKV(w io.Writer, label, value string) {
	fmt.Fprintf(w, "  %-10s %s\n", label+":", value)
}

// printStatus はゲームの現在状態を書き出す。
func printStatus(st *api.GameState) {
	outputMu.Lock()
	defer outputMu.Unlock()

	printSection(Stdout, "== ゲーム状況 ==")
	printKV(Stdout, "日付", st.Date.String())
	fmt.Fprintf(Stdout, "  %-10s %d / %d 位\n", "ターン:", st.Turn, st.Rank)
	if c := st.Company; c != nil {
		printKV(Stdout, "会社", c.Name)
		printKV(Stdout, "現金", formatYen(c.Cash))
		printKV(Stdout, "負債", formatYen(c.Debt))
		printKV(Stdout, "純資産", formatYen(c.NetAssets()))
		fmt.Fprintf(Stdout, "  %-10s %d / %d 社\n", "路線:", len(c.Lines), st.TotalCompanies)
		printKV(Stdout, "評判", fmt.Sprintf("%d", c.Reputation))
	} else {
		colorWarn.Fprintln(Stdout, "  プレイヤー会社がありません")
	}
}

// newTable は CLI 用のテーブルを生成する。
func newTable(header []string) *tablewriter.Table {
	t := tablewriter.NewWriter(Stdout)
	t.SetHeader(header)
	t.SetAutoFormatHeaders(false)
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetBorder(false)
	return t
}

// printLines は路線一覧をテーブルで書き出す。
func printLines(lines []core.Line, cities map[string]*core.City) {
	outputMu.Lock()
	defer outputMu.Unlock()

	if len(lines) == 0 {
		colorWarn.Fprintln(Stdout, "  路線がありません")
		return
	}
	t := newTable([]string{"ID", "名称", "区間", "距離", "運賃", "月次利益", "営業係数", "状態"})
	for i := range lines {
		l := &lines[i]
		t.Append([]string{
			l.ID,
			l.DisplayName,
			cityName(cities, l.From) + " - " + cityName(cities, l.To),
			fmt.Sprintf("%.1fkm", l.LengthKm),
			fmt.Sprintf("%d円", l.Fare),
			formatYen(l.Profit),
			fmt.Sprintf("%.1f", l.OperatingRatio),
			l.Health(),
		})
	}
	t.Render()
}

// printLine は 1 路線の詳細を書き出す。
func printLine(l *core.Line, cities map[string]*core.City) {
	outputMu.Lock()
	defer outputMu.Unlock()

	printSection(Stdout, "== 路線 "+l.ID+" ==")
	printKV(Stdout, "名称", l.DisplayName)
	printKV(Stdout, "区間", cityName(cities, l.From)+" - "+cityName(cities, l.To))
	printKV(Stdout, "距離", fmt.Sprintf("%.1fkm", l.LengthKm))
	printKV(Stdout, "軌間", fmt.Sprintf("%dmm", l.Gauge))
	printKV(Stdout, "複線", boolJP(l.Double))
	printKV(Stdout, "電化", boolJP(l.Electrified))
	printKV(Stdout, "運賃", fmt.Sprintf("%d円", l.Fare))
	printKV(Stdout, "本数", fmt.Sprintf("%d本/日", l.TrainsPerDay))
	printKV(Stdout, "月次利益", formatYen(l.Profit))
	printKV(Stdout, "営業係数", fmt.Sprintf("%.1f", l.OperatingRatio))
	printKV(Stdout, "状態", l.Health())
	printKV(Stdout, "連続赤字", fmt.Sprintf("%dターン", l.ConsecutiveDeficitTurns))
}

// printCompanies は会社一覧をテーブルで書き出す。
func printCompanies(companies []core.Company, playerID string) {
	outputMu.Lock()
	defer outputMu.Unlock()

	t := newTable([]string{"ID", "名称", "現金", "負債", "株価", "路線", "区分"})
	for i := range companies {
		c := &companies[i]
		kind := ""
		switch {
		case c.ID == playerID:
			kind = "自社"
		case c.ParentID == playerID:
			kind = "子会社"
		case c.IsNational:
			kind = "国鉄"
		}
		t.Append([]string{
			c.ID,
			c.Name,
			formatYen(c.Cash),
			formatYen(c.Debt),
			fmt.Sprintf("%d円", c.StockPrice),
			fmt.Sprintf("%d", len(c.Lines)),
			kind,
		})
	}
	t.Render()
}

// printCities は都市一覧をテーブルで書き出す。
func printCities(cities []core.City) {
	outputMu.Lock()
	defer outputMu.Unlock()

	t := newTable([]string{"ID", "名称", "人口", "需要", "開発度", "地価", "所有者"})
	for i := range cities {
		c := &cities[i]
		t.Append([]string{
			c.ID,
			c.Name,
			core.FormatNumber(int64(c.Population)),
			fmt.Sprintf("%d", c.Demand),
			fmt.Sprintf("%.2f", c.Developed),
			fmt.Sprintf("%d", c.LandValue),
			c.OwnerID,
		})
	}
	t.Render()
}

// printGates は敷設可否の判定結果を書き出す。
func printGates(r *api.CanBuildResult) {
	outputMu.Lock()
	defer outputMu.Unlock()

	if r.Can {
		colorSuccess.Fprintln(Stdout, "敷設可能です。")
	} else {
		colorDanger.Fprintln(Stdout, "敷設できません。")
	}
	printKV(Stdout, "概算費用", formatYen(r.Cost))
	printKV(Stdout, "工期", fmt.Sprintf("%dヶ月", r.Months))

	t := newTable([]string{"ゲート", "判定", "理由"})
	for _, g := range r.Gates {
		t.Append([]string{g.Kind, g.Status, g.Reason})
	}
	t.Render()
}

// printEvents はイベント一覧を重要度別の色で書き出す。
func printEvents(events []core.Event) {
	outputMu.Lock()
	defer outputMu.Unlock()
	printEventsLocked(events)
}

// printEventsLocked は outputMu 保持前提でイベントを書き出す。
func printEventsLocked(events []core.Event) {
	for i := range events {
		printEventLocked(&events[i])
	}
}

// printEventLocked は 1 イベントを書き出す。
func printEventLocked(e *core.Event) {
	c := colorFor(e.Severity)
	prefix := e.Date.String()
	c.Fprintf(Stdout, "[%s] %s\n", prefix, e.Message)
}

// colorFor は重要度に対応する色を返す。
func colorFor(severity string) *color.Color {
	switch severity {
	case core.SeveritySuccess:
		return colorSuccess
	case core.SeverityWarning:
		return colorWarn
	case core.SeverityDanger:
		return colorDanger
	default:
		return colorInfo
	}
}

// PrintError はエラーを赤字で stderr へ書き出す。
func PrintError(err error) {
	if err == nil {
		return
	}
	outputMu.Lock()
	defer outputMu.Unlock()
	colorDanger.Fprintf(Stderr, "エラー: %v\n", err)
}

// printSuccess は成功メッセージを緑字で書き出す。
func printSuccess(format string, args ...any) {
	outputMu.Lock()
	defer outputMu.Unlock()
	colorSuccess.Fprintf(Stdout, format+"\n", args...)
}

// printInfo は通常メッセージを書き出す。
func printInfo(format string, args ...any) {
	outputMu.Lock()
	defer outputMu.Unlock()
	fmt.Fprintf(Stdout, format+"\n", args...)
}

// printGroups は [][]string をテーブルで書き出す（汎用）。
func printTable(header []string, rows [][]string) {
	outputMu.Lock()
	defer outputMu.Unlock()
	t := newTable(header)
	for _, r := range rows {
		t.Append(r)
	}
	t.Render()
}

// cityName は都市 ID から名称を引く。未登録なら ID をそのまま返す。
func cityName(cities map[string]*core.City, id string) string {
	if c, ok := cities[id]; ok && c != nil {
		return c.Name
	}
	return id
}

// boolJP は真偽値を日本語にする。
func boolJP(b bool) string {
	if b {
		return "はい"
	}
	return "いいえ"
}
