package cli

// HandlerFunc は 1 コマンドの実装。
type HandlerFunc func(*Session, *Command) error

// handlers は正式動詞 → ハンドラの登録表。別名は Execute で解決される。
var handlers = map[string]HandlerFunc{
	// セッション
	"start": handleStart,
	"load":  handleLoad,
	"list":  handleList,
	"quit":  handleQuit,

	// 状態
	"status":    handleStatus,
	"lines":     handleLines,
	"line":      handleLine,
	"companies": handleCompanies,
	"cities":    handleCities,
	"can-build": handleCanBuild,

	// 建設・運用
	"build":          handleBuild,
	"build-facility": handleBuildFacility,
	"set-fare":       handleSetFare,
	"abandon-line":   handleAbandonLine,

	// M&A
	"buy-shares": handleBuyShares,
	"tob":        handleTOB,
	"merge":      handleMerge,

	// ターン
	"et":   handleEndTurn,
	"turn": handleTurnN,

	// ヘルプ
	"help": handleHelp,
}
