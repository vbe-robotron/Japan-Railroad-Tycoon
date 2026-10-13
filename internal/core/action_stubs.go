package core

// このファイルは Phase 1-4 時点での一時スタブ。
// 本実装済みのアクション（EndTurnAction）はここから削除済み。
// Phase 1-5 以降、各アクションの本実装が完了したら、
// 対応するメソッドをこのファイルから削除する。
//
// Validate は状態を変更せず実行可否を検証し、Apply は状態を変更して
// 発生したイベントを返す。ここではすべて無条件に成功する空実装とする。

// ────────────────────────────────────────────
// 建設系（5種）
// ────────────────────────────────────────────

// Validate は路線敷設の可否を検証する（スタブ）。
func (a *BuildLineAction) Validate(g *Game) error { return nil }

// Apply は路線を敷設する（スタブ）。
func (a *BuildLineAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は中間駅追加の可否を検証する（スタブ）。
func (a *BuildStationAction) Validate(g *Game) error { return nil }

// Apply は中間駅を追加する（スタブ）。
func (a *BuildStationAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は施設建設の可否を検証する（スタブ）。
func (a *BuildFacilityAction) Validate(g *Game) error { return nil }

// Apply は施設を建設する（スタブ）。
func (a *BuildFacilityAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は車庫・工場・発電所建設の可否を検証する（スタブ）。
func (a *BuildDepotAction) Validate(g *Game) error { return nil }

// Apply は車庫・工場・発電所を建設する（スタブ）。
func (a *BuildDepotAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は電化の可否を検証する（スタブ）。
func (a *ElectrifyAction) Validate(g *Game) error { return nil }

// Apply は路線を電化する（スタブ）。
func (a *ElectrifyAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// ────────────────────────────────────────────
// 運用系（3種）
// ────────────────────────────────────────────

// Validate は車両購入の可否を検証する（スタブ）。
func (a *PurchaseRollingStockAction) Validate(g *Game) error { return nil }

// Apply は車両を購入する（スタブ）。
func (a *PurchaseRollingStockAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は運賃変更の可否を検証する（スタブ）。
func (a *SetFareAction) Validate(g *Game) error { return nil }

// Apply は運賃を変更する（スタブ）。
func (a *SetFareAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate はダイヤ変更の可否を検証する（スタブ）。
func (a *SetScheduleAction) Validate(g *Game) error { return nil }

// Apply はダイヤを変更する（スタブ）。
func (a *SetScheduleAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// ────────────────────────────────────────────
// 財務系（4種）
// ────────────────────────────────────────────

// Validate は社債発行の可否を検証する（スタブ）。
func (a *IssueBondsAction) Validate(g *Game) error { return nil }

// Apply は社債を発行する（スタブ）。
func (a *IssueBondsAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は増資の可否を検証する（スタブ）。
func (a *IssueSharesAction) Validate(g *Game) error { return nil }

// Apply は増資する（スタブ）。
func (a *IssueSharesAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は配当支払いの可否を検証する（スタブ）。
func (a *PayDividendAction) Validate(g *Game) error { return nil }

// Apply は配当を支払う（スタブ）。
func (a *PayDividendAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は借金返済の可否を検証する（スタブ）。
func (a *RepayDebtAction) Validate(g *Game) error { return nil }

// Apply は借金を返済する（スタブ）。
func (a *RepayDebtAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// ────────────────────────────────────────────
// M&A系（5種）
// ────────────────────────────────────────────

// Validate は他社株取得の可否を検証する（スタブ）。
func (a *BuySharesAction) Validate(g *Game) error { return nil }

// Apply は他社株を取得する（スタブ）。
func (a *BuySharesAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は TOB の可否を検証する（スタブ）。
func (a *TenderOfferAction) Validate(g *Game) error { return nil }

// Apply は TOB を実行する（スタブ）。
func (a *TenderOfferAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は合併の可否を検証する（スタブ）。
func (a *MergeAction) Validate(g *Game) error { return nil }

// Apply は他社と合併する（スタブ）。
func (a *MergeAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は提携の可否を検証する（スタブ）。
func (a *AllianceAction) Validate(g *Game) error { return nil }

// Apply は他社と提携する（スタブ）。
func (a *AllianceAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は買収防衛策の可否を検証する（スタブ）。
func (a *DefendTakeoverAction) Validate(g *Game) error { return nil }

// Apply は買収防衛策を講じる（スタブ）。
func (a *DefendTakeoverAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// ────────────────────────────────────────────
// 開発系（2種）
// ────────────────────────────────────────────

// Validate は用地取得の可否を検証する（スタブ）。
func (a *AcquireLandAction) Validate(g *Game) error { return nil }

// Apply は用地を取得する（スタブ）。
func (a *AcquireLandAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は政治工作の可否を検証する（スタブ）。
func (a *LobbyAction) Validate(g *Game) error { return nil }

// Apply は政治家に工作する（スタブ）。
func (a *LobbyAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// ────────────────────────────────────────────
// 政治系（3種）
// ────────────────────────────────────────────

// Validate は国有化の可否を検証する（スタブ）。
func (a *NationalizeAction) Validate(g *Game) error { return nil }

// Apply は対象会社を国有化する（スタブ）。
func (a *NationalizeAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は国鉄分割民営化の可否を検証する（スタブ）。
func (a *SplitJnrAction) Validate(g *Game) error { return nil }

// Apply は国鉄を分割民営化する（スタブ）。
func (a *SplitJnrAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は補助金受諾の可否を検証する（スタブ）。
func (a *AcceptSubsidyAction) Validate(g *Game) error { return nil }

// Apply は補助金を受諾する（スタブ）。
func (a *AcceptSubsidyAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// ────────────────────────────────────────────
// 廃線系（2種）
// ────────────────────────────────────────────

// Validate は廃線の可否を検証する（スタブ）。
func (a *AbandonLineAction) Validate(g *Game) error { return nil }

// Apply は路線を廃止する（スタブ）。
func (a *AbandonLineAction) Apply(g *Game) ([]Event, error) { return nil, nil }

// Validate は廃線復活の可否を検証する（スタブ）。
func (a *RestoreAbandonedLineAction) Validate(g *Game) error { return nil }

// Apply は廃線を復活させる（スタブ）。
func (a *RestoreAbandonedLineAction) Apply(g *Game) ([]Event, error) { return nil, nil }
