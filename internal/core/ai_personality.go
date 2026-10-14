package core

// このファイルは AI の性格付けを提供する。
//
// 各パラメータは 0.0-1.0 の範囲を取り、Phase 8-2 以降の意思決定
// （買収・拡大・新技術・不動産・政治工作）で重みとして利用する。

// Personality は AI の性格を表す。
type Personality struct {
	Aggression   float64 `json:"aggression"`   // 0-1: 買収・拡大志向
	Conservatism float64 `json:"conservatism"` // 0-1: 安全志向
	Innovation   float64 `json:"innovation"`   // 0-1: 新技術導入
	RealEstate   float64 `json:"real_estate"`  // 0-1: 不動産開発志向
	Politics     float64 `json:"politics"`     // 0-1: 政治工作志向
}

// DefaultPersonality は会社IDから性格を返す。
//
// 史実の経営者像を反映する。未知の会社は中庸（すべて 0.5）を返す。
func DefaultPersonality(companyID string) Personality {
	switch companyID {
	case "hankyu":
		return Personality{Aggression: 0.7, Conservatism: 0.4, Innovation: 0.8, RealEstate: 0.9, Politics: 0.5}
	case "tokyu":
		return Personality{Aggression: 0.9, Conservatism: 0.3, Innovation: 0.7, RealEstate: 0.8, Politics: 0.6}
	case "kintetsu":
		return Personality{Aggression: 0.8, Conservatism: 0.5, Innovation: 0.6, RealEstate: 0.6, Politics: 0.7}
	case "hanshin":
		return Personality{Aggression: 0.4, Conservatism: 0.7, Innovation: 0.4, RealEstate: 0.5, Politics: 0.4}
	case "meitetsu":
		return Personality{Aggression: 0.6, Conservatism: 0.5, Innovation: 0.5, RealEstate: 0.6, Politics: 0.6}
	case "jnr":
		return Personality{Aggression: 0.5, Conservatism: 0.8, Innovation: 0.6, RealEstate: 0.1, Politics: 0.9}
	default:
		return Personality{Aggression: 0.5, Conservatism: 0.5, Innovation: 0.5, RealEstate: 0.5, Politics: 0.5}
	}
}

// InitAIs は Game の AIs マップを初期化する。
//
// 既に AI が割り当てられている会社は上書きしない（冪等）。
// プレイヤー会社は HistoryIgnore、それ以外は HistoryStrict とする。
func (g *Game) InitAIs() {
	if g.AIs == nil {
		g.AIs = map[string]*AI{}
	}
	for id, c := range g.Companies {
		if _, ok := g.AIs[id]; ok {
			continue
		}
		mode := HistoryStrict
		if c.IsPlayer {
			mode = HistoryIgnore
		}
		g.AIs[id] = &AI{
			CompanyID:   id,
			Personality: DefaultPersonality(id),
			HistoryMode: mode,
		}
	}
}
