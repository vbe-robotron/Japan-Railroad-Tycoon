package core

// このファイルは都市の需要再計算ヘルパーを実装する。
//
// 需要は人口・開発度（Game.Demand）に加え、接続路線数による
// ボーナスを受ける。路線敷設や廃線のたびに再計算する。

// recalcCityDemand は都市の需要を再計算する。
//
// 需要は以下に依存:
//   - 人口
//   - 開発度
//   - 接続路線数
func (g *Game) recalcCityDemand(c *City) {
	if c == nil {
		return
	}
	base := g.Demand(c)

	// 接続路線数ボーナス
	connections := g.countConnections(c.ID)
	bonus := 1.0 + float64(connections)*0.1

	c.Demand = int(float64(base) * bonus)
}

// countConnections は指定都市に接続する路線数を返す。
//
// 全会社の路線を対象とする。From/To のいずれかが cityID と一致する
// 路線を数える。
func (g *Game) countConnections(cityID string) int {
	n := 0
	for _, c := range g.Companies {
		for _, l := range c.Lines {
			if l.From == cityID || l.To == cityID {
				n++
			}
		}
	}
	return n
}
