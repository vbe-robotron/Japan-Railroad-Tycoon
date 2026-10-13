// cmd/play/main.go
//
// Phase 1-8-1 完了時点で、core を直接叩いて遊ぶための一時的なコマンド。
// Phase 6 で本格的な CLI が完成したら、このコマンドは削除する。
package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"

    "github.com/you/jrt/internal/core"
)

func main() {
    g := setupGame()
    printBanner(g)

    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Print("jrt> ")
        if !scanner.Scan() {
            break
        }
        line := strings.TrimSpace(scanner.Text())
        if line == "" {
            continue
        }
        if !handleCommand(g, line) {
            break
        }
    }
    fmt.Println("\nさようなら。")
}

// setupGame は初期状態のゲームを作る。
//
// 都市には実座標（Lat/Lon）を設定する。
func setupGame() *core.Game {
    g := core.NewGame(1907, "hankyu")
    g.Date = core.Date{Year: 1907, Month: 10}

    // 都市（実座標付き）
    cities := []*core.City{
        {ID: "umeda",      Name: "梅田",   Prefecture: "大阪府",
         Lat: 34.7055, Lon: 135.4983, Population: 120000, LandValue: 60},
        {ID: "takarazuka", Name: "宝塚",   Prefecture: "兵庫県",
         Lat: 34.7995, Lon: 135.3468, Population: 30000,  LandValue: 20},
        {ID: "mino",       Name: "箕面",   Prefecture: "大阪府",
         Lat: 34.8269, Lon: 135.4705, Population: 20000,  LandValue: 15},
        {ID: "juso",       Name: "十三",   Prefecture: "大阪府",
         Lat: 34.7235, Lon: 135.4843, Population: 40000,  LandValue: 40},
        {ID: "shonai",     Name: "庄内",   Prefecture: "大阪府",
         Lat: 34.7516, Lon: 135.4702, Population: 15000,  LandValue: 25},
        {ID: "kobe",       Name: "神戸",   Prefecture: "兵庫県",
         Lat: 34.6901, Lon: 135.1955, Population: 400000, LandValue: 80},
        {ID: "kyoto",      Name: "京都",   Prefecture: "京都府",
         Lat: 34.9858, Lon: 135.7588, Population: 350000, LandValue: 75},
    }
    for _, c := range cities {
        g.AddCity(c)
    }

    // 阪急会社
    hankyu := core.NewCompany("hankyu", "hankyu", "阪急電鉄")
    hankyu.LegalName = "箕面有馬電気軌道株式会社"
    hankyu.FoundedYear = 1907
    hankyu.Cash = 9_000_000  // 900万円（Phase 1-7.6 の推奨値）
    hankyu.Region = "kansai"
    hankyu.IsPlayer = true
    g.AddCompany(hankyu)

    // 競合他社（AIではなくダミー）
    hanshin := core.NewCompany("hanshin", "hanshin", "阪神電鉄")
    hanshin.Cash = 5_000_000
    hanshin.Region = "kansai"
    g.AddCompany(hanshin)

    return g
}

func printBanner(g *core.Game) {
    c := g.PlayerCompany()
    fmt.Println("=== Japan Railroad Tycoon - Play (Phase 1-8-2) ===")
    fmt.Printf("%s 設立 (%d年)\n", c.Name, c.FoundedYear)
    fmt.Printf("  資金: %s\n", core.FormatYen(c.Cash))
    fmt.Println()
    fmt.Println("コマンド:")
    fmt.Println("  status              自社の概要")
    fmt.Println("  cities              都市一覧")
    fmt.Println("  can-build <from> <to>   敷設可否チェック")
    fmt.Println("  build <from> <to>       路線敷設 [--gauge N] [--double]")
    fmt.Println("  lines               路線一覧（収益付き）")
    fmt.Println("  line <id>           路線詳細")
    fmt.Println("  set-fare <id> <円>  運賃変更（±50%以内）")
    fmt.Println("  build-facility <city> <kind> [--scale N]  施設建設（開発度上昇）")
    fmt.Println("  et                  1ターン進む")
    fmt.Println("  turn <N>            Nターン進む")
    fmt.Println("  help                ヘルプ")
    fmt.Println("  quit                終了")
    fmt.Println()
}

func handleCommand(g *core.Game, line string) bool {
    parts := strings.Fields(line)
    if len(parts) == 0 {
        return true
    }

    switch parts[0] {
    case "status", "st":
        cmdStatus(g)
    case "cities":
        cmdCities(g)
    case "can-build":
        if len(parts) < 3 {
            fmt.Println("使い方: can-build <from> <to>")
            return true
        }
        cmdCanBuild(g, parts[1], parts[2])
    case "build":
        if len(parts) < 3 {
            fmt.Println("使い方: build <from> <to> [--gauge N] [--double]")
            return true
        }
        cmdBuild(g, parts)
    case "lines", "ls":
        cmdLines(g)
    case "line":
        if len(parts) < 2 {
            fmt.Println("使い方: line <id>")
            return true
        }
        cmdLineDetail(g, parts[1])
    case "set-fare", "fare":
        if len(parts) < 3 {
            fmt.Println("使い方: set-fare <line_id> <fare>")
            return true
        }
        fare, err := strconv.Atoi(parts[2])
        if err != nil {
            fmt.Println("運賃は整数で指定してください")
            return true
        }
        cmdSetFare(g, parts[1], fare)
    case "build-facility", "bf":
        if len(parts) < 3 {
            fmt.Println("使い方: build-facility <city> <kind> [--scale N]")
            fmt.Println("  kind: department_store | housing | amusement | hotel | office")
            return true
        }
        cmdBuildFacility(g, parts)
    case "et":
        cmdEndTurn(g, 1)
    case "turn":
        if len(parts) < 2 {
            fmt.Println("使い方: turn <N>")
            return true
        }
        n, err := strconv.Atoi(parts[1])
        if err != nil || n < 1 {
            fmt.Println("N は正の整数で")
            return true
        }
        cmdEndTurn(g, n)
    case "help":
        cmdHelp()
    case "quit", "exit", "q":
        return false
    default:
        fmt.Printf("不明なコマンド: %s（help で一覧）\n", parts[0])
    }
    return true
}

func cmdStatus(g *core.Game) {
    c := g.PlayerCompany()
    fmt.Printf("[%s] %s\n", g.Date.String(), c.Name)
    fmt.Printf("  資金: %s  路線: %d本\n",
        core.FormatYen(c.Cash), len(c.Lines))

    if len(c.Lines) > 0 {
        monthly := c.MonthlyProfit(g)
        annual := c.AnnualProfit(g)
        fmt.Printf("  月次利益: %s  年次利益: %s\n",
            core.FormatYen(monthly), core.FormatYen(annual))
    }
}

func cmdCities(g *core.Game) {
    fmt.Println("都市:")
    fmt.Println("  ID            名前      人口      需要")
    for _, c := range g.Cities {
        fmt.Printf("  %-12s  %-6s  %7d   %5d\n",
            c.ID, c.Name, c.Population, c.Demand)
    }
}

func cmdCanBuild(g *core.Game, from, to string) {
    result := g.CanBuildLine(from, to, 1067, false)

    fmt.Printf("敷設可否: %s → %s\n", from, to)
    if result.Can {
        fmt.Println("  → 敷設可能")
    } else {
        fmt.Println("  → 敷設不可")
    }
    fmt.Printf("  コスト: %s\n", core.FormatYen(result.Cost))
    fmt.Printf("  工期: %dヶ月\n", result.Months)
    for _, gate := range result.Gates {
        mark := "●"
        switch gate.Status {
        case core.GatePass:
            mark = "●"
        case core.GateConditional:
            mark = "◐"
        case core.GateFail:
            mark = "✗"
        }
        fmt.Printf("  %s [%s] %s\n", mark, gate.Kind, gate.Reason)
    }
}

func cmdBuild(g *core.Game, parts []string) {
    from, to := parts[1], parts[2]
    gauge := 1067
    double := false

    for i := 3; i < len(parts); i++ {
        switch parts[i] {
        case "--double":
            double = true
        case "--gauge":
            if i+1 < len(parts) {
                v, _ := strconv.Atoi(parts[i+1])
                gauge = v
                i++
            }
        }
    }

    a := &core.BuildLineAction{
        From: from, To: to, Gauge: gauge, Double: double,
    }
    events, err := a.Apply(g)
    if err != nil {
        fmt.Printf("✗ %v\n", err)
        return
    }
    printEvents(events)
}

func cmdLines(g *core.Game) {
    c := g.PlayerCompany()
    if len(c.Lines) == 0 {
        fmt.Println("路線がありません")
        return
    }
    fmt.Printf("%-10s %-20s %8s %8s %10s %8s\n",
        "ID", "区間", "距離", "運賃", "月次利益", "営業係数")
    for _, l := range c.Lines {
        from := g.Cities[l.From]
        to := g.Cities[l.To]
        fmt.Printf("%-10s %-8s〜%-8s %6.1fkm %6d円 %10s %8.1f\n",
            shortID(l.ID),
            from.Name, to.Name,
            l.LengthKm, l.Fare,
            core.FormatYen(l.Profit),
            l.OperatingRatio)
    }
    fmt.Printf("\n月次利益合計: %s\n", core.FormatYen(c.MonthlyProfit(g)))
}

func cmdLineDetail(g *core.Game, id string) {
    c := g.PlayerCompany()
    for _, l := range c.Lines {
        if shortID(l.ID) == id || l.ID == id {
            fmt.Printf("路線: %s\n", l.DisplayName)
            fmt.Printf("  区間: %s → %s\n", l.From, l.To)
            fmt.Printf("  距離: %.1fkm\n", l.LengthKm)
            fmt.Printf("  軌間: %dmm\n", l.Gauge)
            fmt.Printf("  複線: %v  電化: %v\n", l.Double, l.Electrified)
            fmt.Printf("  運賃: %d円  本数: %d本/日\n", l.Fare, l.TrainsPerDay)
            fmt.Printf("  月次利益: %s\n", core.FormatYen(l.Profit))
            fmt.Printf("  営業係数: %.1f\n", l.OperatingRatio)
            fmt.Printf("  状態: %s\n", l.Health())
            fmt.Printf("  連続赤字: %dヶ月  累積損失: %s\n",
                l.ConsecutiveDeficitTurns, core.FormatYen(l.AccumulatedLoss))
            return
        }
    }
    fmt.Printf("路線が見つかりません: %s\n", id)
}

// cmdSetFare は運賃変更アクションを適用する。
//
// line_id は完全な ID か、lines 表示の短縮 ID（先頭 8 文字）を受け付ける。
func cmdSetFare(g *core.Game, lineID string, fare int) {
    c := g.PlayerCompany()
    full := lineID
    for _, l := range c.Lines {
        if shortID(l.ID) == lineID || l.ID == lineID {
            full = l.ID
            break
        }
    }

    a := &core.SetFareAction{LineID: full, Fare: fare}
    events, err := a.Apply(g)
    if err != nil {
        fmt.Printf("✗ %v\n", err)
        return
    }
    printEvents(events)
}

// cmdBuildFacility は施設建設アクションを適用する。
//
// kind は department_store | housing | amusement | hotel | office。
// --scale N で規模（1〜5、既定値 1）を指定する。
func cmdBuildFacility(g *core.Game, parts []string) {
    cityID, kind := parts[1], parts[2]
    scale := 1
    for i := 3; i < len(parts); i++ {
        if parts[i] == "--scale" && i+1 < len(parts) {
            scale, _ = strconv.Atoi(parts[i+1])
            i++
        }
    }
    a := &core.BuildFacilityAction{CityID: cityID, Kind: kind, Scale: scale}
    events, err := a.Apply(g)
    if err != nil {
        fmt.Printf("✗ %v\n", err)
        return
    }
    printEvents(events)
}

func cmdEndTurn(g *core.Game, n int) {
    for i := 0; i < n; i++ {
        events := g.EndTurn()
        printEvents(events)
    }
    c := g.PlayerCompany()
    fmt.Printf("  → 資金: %s\n", core.FormatYen(c.Cash))
}

func cmdHelp() {
    fmt.Println("コマンド:")
    fmt.Println("  status              自社の概要")
    fmt.Println("  cities              都市一覧")
    fmt.Println("  can-build <from> <to>   敷設可否")
    fmt.Println("  build <from> <to>       路線敷設")
    fmt.Println("  lines               路線一覧（収益付き）")
    fmt.Println("  line <id>           路線詳細")
    fmt.Println("  set-fare <id> <円>  運賃変更（±50%以内）")
    fmt.Println("  build-facility <city> <kind> [--scale N]  施設建設（開発度上昇）")
    fmt.Println("  et                  1ターン進む")
    fmt.Println("  turn <N>            Nターン進む")
    fmt.Println("  help                ヘルプ")
    fmt.Println("  quit                終了")
}

func printEvents(events []core.Event) {
    for _, e := range events {
        icon := "  "
        switch e.Severity {
        case core.SeveritySuccess:
            icon = "✓ "
        case core.SeverityWarning:
            icon = "⚠ "
        case core.SeverityDanger:
            icon = "✗ "
        }
        fmt.Printf("[%s] %s%s\n", e.Date.String(), icon, e.Message)
    }
}

func shortID(id string) string {
    if len(id) <= 8 {
        return id
    }
    return id[:8]
}
