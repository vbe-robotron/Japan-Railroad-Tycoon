package core

import "fmt"

// Event はゲーム内で発生した出来事を表す。
//
// 全てのアクションは []Event を返し、WebSocket で配信され、DB に保存され、
// CLI では Message をそのまま表示する。GUI/AI は Payload を解釈する。
type Event struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Date      Date           `json:"date"`
	Turn      int            `json:"turn"`
	Severity  string         `json:"severity"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload,omitempty"`
	CompanyID string         `json:"company_id,omitempty"`
	CityID    string         `json:"city_id,omitempty"`
	LineID    string         `json:"line_id,omitempty"`
}

// NewEvent はイベントを生成する。
//
// ID は決定論的 ID 生成器から払い出すため、同じ seed と同じアクション列から
// 常に同じイベント列が得られる。Date/Turn は現在のゲーム状態から設定する。
func (g *Game) NewEvent(eventType, severity, message string) Event {
	return Event{
		ID:       g.IDGen.Next(),
		Type:     eventType,
		Date:     g.Date,
		Turn:     g.Turn,
		Severity: severity,
		Message:  message,
	}
}

// With は Payload にキーと値を追加した新しい Event を返す。
//
// 値レシーバだが Payload は map（参照型）であるため、そのまま書き込むと
// 元の Event と共有されてしまう。ここでは map を毎回コピーすることで
// イミュータブルを保証する。チェーンしても元の Event は変化しない。
func (e Event) With(k string, v any) Event {
	newPayload := make(map[string]any, len(e.Payload)+1)
	for key, val := range e.Payload {
		newPayload[key] = val
	}
	newPayload[k] = v
	e.Payload = newPayload
	return e
}

// ForCompany は CompanyID を設定した新しい Event を返す。
func (e Event) ForCompany(id string) Event {
	e.CompanyID = id
	return e
}

// ForLine は LineID を設定した新しい Event を返す。
func (e Event) ForLine(id string) Event {
	e.LineID = id
	return e
}

// ForCity は CityID を設定した新しい Event を返す。
func (e Event) ForCity(id string) Event {
	e.CityID = id
	return e
}

// Info は重要度 info のイベントを生成する。
func (g *Game) Info(t, msg string) Event {
	return g.NewEvent(t, SeverityInfo, msg)
}

// Success は重要度 success のイベントを生成する。
func (g *Game) Success(t, msg string) Event {
	return g.NewEvent(t, SeveritySuccess, msg)
}

// Warn は重要度 warning のイベントを生成する。
func (g *Game) Warn(t, msg string) Event {
	return g.NewEvent(t, SeverityWarning, msg)
}

// Danger は重要度 danger のイベントを生成する。
func (g *Game) Danger(t, msg string) Event {
	return g.NewEvent(t, SeverityDanger, msg)
}

// Evt はテンプレートからイベントを生成する。
//
// format と args は fmt.Sprintf に渡され、その結果が Message になる。
// severity は IsValidSeverity を満たす必要がある（検証は呼び出し側の責務）。
func (g *Game) Evt(t, sev, format string, args ...any) Event {
	return g.NewEvent(t, sev, fmt.Sprintf(format, args...))
}
