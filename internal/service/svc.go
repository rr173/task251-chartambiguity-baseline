// Package service 是编排层，将 store 与各业务包（layer/mapping/checker/spec）
// 组合为跨实体的业务闭环：导入语义 → 声明编码 → 复核歧义 → 豁免 → 发布规范。
package service

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"task251-chartambiguity/internal/checker"
	"task251-chartambiguity/internal/layer"
	"task251-chartambiguity/internal/mapping"
	"task251-chartambiguity/internal/model"
	"task251-chartambiguity/internal/spec"
	"task251-chartambiguity/internal/store"
)

// Service 持有存储与全部业务规则，对外提供高层操作。
type Service struct {
	store *store.Store
}

// New 构造 Service。
func New(s *store.Store) *Service { return &Service{store: s} }

// Store 暴露底层存储，供需要在事务内组合多实体读取的调用方使用。
func (svc *Service) Store() *store.Store { return svc.store }

// ---- 请求负载 ----

// LayerInput / AxisInput / LegendInput / VariableInput 为导入语义的单项输入。
type LayerInput struct {
	Name      string `json:"name"`
	LayerType string `json:"layer_type"`
	ZOrder    int    `json:"z_order"`
	Visible   bool   `json:"visible"`
}
type AxisInput struct {
	Name        string `json:"name"`
	Variable    string `json:"variable"`
	Unit        string `json:"unit"`
	Orientation string `json:"orientation"`
}
type LegendInput struct {
	Channel        string `json:"channel"`
	Label          string `json:"label"`
	Token          string `json:"token"`
	CoversVariable string `json:"covers_variable"`
}
type VariableInput struct {
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	Description string `json:"description"`
}
type ImportPayload struct {
	Layers    []LayerInput    `json:"layers"`
	Axes      []AxisInput     `json:"axes"`
	Legends   []LegendInput   `json:"legends"`
	Variables []VariableInput `json:"variables"`
}

// ExceptionInput 为新增例外的输入。
type ExceptionInput struct {
	Kind           string `json:"kind"`
	TargetChannel  string `json:"target_channel"`
	TargetToken    string `json:"target_token"`
	TargetVariable string `json:"target_variable"`
	Reason         string `json:"reason"`
}

// ---- 图形稿 ----

// CreateFigure 创建一篇图形稿（状态 importing）。
func (svc *Service) CreateFigure(title string) (*model.Figure, error) {
	if strings.TrimSpace(title) == "" {
		return nil, model.ErrInvalidArgument
	}
	now := time.Now().Unix()
	f := &model.Figure{
		ID:        model.NewID("fig"),
		Title:     title,
		Status:    model.FigureStatusImporting,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := svc.store.CreateFigure(f); err != nil {
		return nil, err
	}
	return f, nil
}

// GetFigure 读取图形稿。
func (svc *Service) GetFigure(id string) (*model.Figure, error) { return svc.store.GetFigure(id) }

// ListFigures 列出图形稿。
func (svc *Service) ListFigures(limit int) ([]model.Figure, error) {
	return svc.store.ListFigures(limit)
}

// ImportSemantics 导入图层/轴/图例/变量语义，并派生指纹与图层摘要，
// 将图形稿状态推进到 pending_review。已冻结图形稿禁止导入。
func (svc *Service) ImportSemantics(figureID string, p ImportPayload) error {
	f, err := svc.store.GetFigure(figureID)
	if err != nil {
		return err
	}
	if !f.CanEdit() {
		return model.ErrFrozen
	}
	// 计算导入语义指纹，相同输入幂等；先比较再写入，避免重复导入制造重复实体。
	fpParts := make([]string, 0, len(p.Layers)+len(p.Axes)+len(p.Variables)+len(p.Legends))
	for _, l := range p.Layers {
		fpParts = append(fpParts, "L:"+l.Name+":"+l.LayerType+":"+strconv.Itoa(l.ZOrder)+":"+strconv.FormatBool(l.Visible))
	}
	for _, a := range p.Axes {
		fpParts = append(fpParts, "A:"+a.Name+":"+a.Variable+":"+a.Unit+":"+a.Orientation)
	}
	for _, v := range p.Variables {
		fpParts = append(fpParts, "V:"+v.Name+":"+v.Unit+":"+v.Description)
	}
	for _, g := range p.Legends {
		fpParts = append(fpParts, "G:"+g.Channel+":"+g.Label+":"+g.Token+":"+g.CoversVariable)
	}
	sort.Strings(fpParts)
	fp := model.Fingerprint(fpParts...)
	// 先完整校验输入，再执行写入，避免后续非法项留下前面已写入的半套语义。
	for _, li := range p.Layers {
		if err := layer.ValidateLayer(model.Layer{Name: li.Name, LayerType: li.LayerType, ZOrder: li.ZOrder}); err != nil {
			return err
		}
	}
	for _, ai := range p.Axes {
		if err := layer.ValidateAxis(model.Axis{Name: ai.Name, Variable: ai.Variable, Unit: ai.Unit, Orientation: ai.Orientation}); err != nil {
			return err
		}
	}
	for _, vi := range p.Variables {
		if err := layer.ValidateVariable(model.Variable{Name: vi.Name}); err != nil {
			return err
		}
	}
	for _, gi := range p.Legends {
		if err := layer.ValidateLegend(model.Legend{Channel: gi.Channel, Token: gi.Token}); err != nil {
			return err
		}
	}
	for _, li := range p.Layers {
		l := model.Layer{ID: model.NewID("lyr"), FigureID: figureID, Name: li.Name, LayerType: li.LayerType, ZOrder: li.ZOrder, Visible: li.Visible}
		if err := svc.store.CreateLayer(&l); err != nil {
			return err
		}
	}
	for _, ai := range p.Axes {
		a := model.Axis{ID: model.NewID("ax"), FigureID: figureID, Name: ai.Name, Variable: ai.Variable, Unit: ai.Unit, Orientation: ai.Orientation}
		if err := svc.store.CreateAxis(&a); err != nil {
			return err
		}
	}
	for _, vi := range p.Variables {
		v := model.Variable{ID: model.NewID("var"), FigureID: figureID, Name: vi.Name, Unit: vi.Unit, Description: vi.Description}
		if err := svc.store.CreateVariable(&v); err != nil {
			return err
		}
	}
	for _, gi := range p.Legends {
		g := model.Legend{ID: model.NewID("lgd"), FigureID: figureID, Channel: gi.Channel, Label: gi.Label, Token: gi.Token, CoversVariable: gi.CoversVariable}
		if err := svc.store.CreateLegend(&g); err != nil {
			return err
		}
	}

	if err := svc.store.SetFigureSourceFP(figureID, fp, len(p.Layers)); err != nil {
		return err
	}
	if err := svc.store.UpdateFigureStatus(figureID, model.FigureStatusPendingReview); err != nil {
		return err
	}
	return nil
}

// DeclareEncoding 声明一条视觉编码（变量经通道以 token 呈现）。
func (svc *Service) DeclareEncoding(figureID, layerID, variable, channel, token string) (*model.VisualEncoding, error) {
	if variable == "" || !model.IsValidChannel(channel) || token == "" {
		return nil, model.ErrInvalidArgument
	}
	f, err := svc.store.GetFigure(figureID)
	if err != nil {
		return nil, err
	}
	if !f.CanEdit() {
		return nil, model.ErrFrozen
	}
	if _, err := svc.store.GetVariable(figureID, variable); err != nil {
		return nil, err
	}
	if layerID != "" {
		layers, err := svc.store.ListLayers(figureID)
		if err != nil {
			return nil, err
		}
		found := false
		for _, l := range layers {
			if l.ID == layerID {
				found = true
				break
			}
		}
		if !found {
			return nil, model.ErrNotFound
		}
	}
	e := &model.VisualEncoding{
		ID:       model.NewID("enc"),
		FigureID: figureID,
		LayerID:  layerID,
		Variable: variable,
		Channel:  channel,
		Token:    token,
		Status:   model.EncodingStatusParsed,
	}
	if err := svc.store.CreateEncoding(e); err != nil {
		return nil, err
	}
	return e, nil
}

// RunCheck 重算变量-通道映射与全部歧义，更新编码状态与图形稿状态。
// 返回本次计算出的歧义集合（含已被豁免者）。
func (svc *Service) RunCheck(figureID string) ([]model.Ambiguity, error) {
	figure, err := svc.store.GetFigure(figureID)
	if err != nil {
		return nil, err
	}
	if !figure.CanEdit() {
		return nil, model.ErrFrozen
	}
	encs, err := svc.store.ListEncodings(figureID)
	if err != nil {
		return nil, err
	}
	axes, err := svc.store.ListAxes(figureID)
	if err != nil {
		return nil, err
	}
	legends, err := svc.store.ListLegends(figureID)
	if err != nil {
		return nil, err
	}
	excs, err := svc.store.ListExceptions(figureID)
	if err != nil {
		return nil, err
	}

	// 重算映射（先清后写，保证幂等）。
	maps := mapping.BuildMappings(encs)
	if err := svc.store.DeleteMappings(figureID); err != nil {
		return nil, err
	}
	for i := range maps {
		if err := svc.store.CreateMapping(&maps[i]); err != nil {
			return nil, err
		}
	}

	ambigs := checker.CheckAll(encs, axes, legends, maps, excs)
	for i := range ambigs {
		ambigs[i].FigureID = figureID
	}

	// 写回歧义（先清后写，保证与当前计算一致）。
	if err := svc.store.DeleteAmbiguities(figureID); err != nil {
		return nil, err
	}
	if err := svc.store.InsertAmbiguities(ambigs); err != nil {
		return nil, err
	}

	// 更新编码级状态。
	if err := svc.applyEncodingStatuses(figureID, encs, ambigs); err != nil {
		return nil, err
	}

	// 更新图形稿状态：有未解决歧义→pending_review，否则 publishable。
	open := 0
	for _, a := range ambigs {
		if a.IsOpen() {
			open++
		}
	}
	newStatus := model.FigureStatusPublishable
	if open > 0 {
		newStatus = model.FigureStatusPendingReview
	}
	if err := svc.store.UpdateFigureStatus(figureID, newStatus); err != nil {
		return nil, err
	}
	return ambigs, nil
}

// applyEncodingStatuses 依据歧义把每条编码标记为 valid / ambiguous / missing_legend。
func (svc *Service) applyEncodingStatuses(figureID string, encs []model.VisualEncoding, ambigs []model.Ambiguity) error {
	// 收集开放歧义涉及的 (channel,token)。
	reuse := map[string]bool{}    // channel+token
	missing := map[string]bool{}  // channel+token
	conflict := map[string]bool{} // variable+channel
	for _, a := range ambigs {
		if !a.IsOpen() {
			continue
		}
		key := a.Channel + "|" + a.Token
		switch a.Type {
		case model.AmbiguityColorReuse, model.AmbiguityShapeReuse:
			reuse[key] = true
		case model.AmbiguityMissingLegend:
			missing[key] = true
		case model.AmbiguityMappingConflict:
			conflict[a.Variables+"|"+a.Channel] = true
		}
	}
	for _, e := range encs {
		key := e.Channel + "|" + e.Token
		status := model.EncodingStatusValid
		if reuse[key] || conflict[e.Variable+"|"+e.Channel] {
			status = model.EncodingStatusAmbiguous
		} else if missing[key] {
			status = model.EncodingStatusMissingLgnd
		}
		if status != e.Status {
			if err := svc.store.SetEncodingStatus(e.ID, status); err != nil {
				return err
			}
		}
	}
	return nil
}

// AddException 登记一条豁免声明并立即重算歧义（使其生效）。
func (svc *Service) AddException(figureID string, in ExceptionInput) (*model.Exception, error) {
	if !model.IsValidExceptionKind(in.Kind) {
		return nil, model.ErrInvalidArgument
	}
	figure, err := svc.store.GetFigure(figureID)
	if err != nil {
		return nil, err
	}
	if !figure.CanEdit() {
		return nil, model.ErrFrozen
	}
	exc := &model.Exception{
		ID:             model.NewID("exc"),
		FigureID:       figureID,
		Kind:           in.Kind,
		TargetChannel:  in.TargetChannel,
		TargetToken:    in.TargetToken,
		TargetVariable: in.TargetVariable,
		Reason:         in.Reason,
		CreatedAt:      time.Now().Unix(),
	}
	if err := svc.store.CreateException(exc); err != nil {
		return nil, err
	}
	if _, err := svc.RunCheck(figureID); err != nil {
		return nil, err
	}
	return exc, nil
}

// CreateSpec 创建一份图规范草稿（要求当前无未解决歧义）。
func (svc *Service) CreateSpec(figureID string) (*model.FigureSpec, error) {
	figure, err := svc.store.GetFigure(figureID)
	if err != nil {
		return nil, err
	}
	if figure.Status == model.FigureStatusFrozen {
		return nil, model.ErrFrozen
	}
	if figure.Status != model.FigureStatusPublishable {
		return nil, model.ErrInvalidStatus
	}
	open, err := svc.store.CountOpenAmbiguities(figureID)
	if err != nil {
		return nil, err
	}
	if open > 0 {
		return nil, model.ErrHasOpenAmbiguity
	}
	return svc.writeSpec(figureID, model.SpecStatusDraft)
}

// PublishSpec 冻结指定规范版本，并将同图形稿其它 frozen 版本置为 superseded，
// 同时把图形稿状态推进到 frozen。要求当前无未解决歧义。
func (svc *Service) PublishSpec(figureID, specID string) (*model.FigureSpec, error) {
	figure, err := svc.store.GetFigure(figureID)
	if err != nil {
		return nil, err
	}
	if figure.Status != model.FigureStatusPublishable {
		return nil, model.ErrInvalidStatus
	}
	open, err := svc.store.CountOpenAmbiguities(figureID)
	if err != nil {
		return nil, err
	}
	if open > 0 {
		return nil, model.ErrHasOpenAmbiguity
	}
	sp, err := svc.store.GetSpec(specID)
	if err != nil {
		return nil, err
	}
	if sp.FigureID != figureID {
		return nil, model.ErrNotFound
	}
	if sp.Status != model.SpecStatusDraft {
		return nil, model.ErrInvalidStatus
	}
	currentSnapshot, err := svc.buildSnapshot(figureID)
	if err != nil {
		return nil, err
	}
	if err := svc.store.UpdateSpecSnapshot(specID, currentSnapshot); err != nil {
		return nil, err
	}
	if err := svc.store.UpdateSpecStatus(specID, model.SpecStatusFrozen); err != nil {
		return nil, err
	}
	if err := svc.store.SupersedeFrozen(figureID, specID); err != nil {
		return nil, err
	}
	if err := svc.store.UpdateFigureStatus(figureID, model.FigureStatusFrozen); err != nil {
		return nil, err
	}
	return svc.store.GetSpec(specID)
}

// writeSpec 构建快照并写入一份规范（状态由 status 指定）。
func (svc *Service) writeSpec(figureID, status string) (*model.FigureSpec, error) {
	snap, err := svc.buildSnapshot(figureID)
	if err != nil {
		return nil, err
	}
	maxV, err := svc.store.MaxSpecVersion(figureID)
	if err != nil {
		return nil, err
	}
	sp := &model.FigureSpec{
		ID:        model.NewID("spec"),
		FigureID:  figureID,
		Version:   spec.NextVersion(maxV),
		Status:    status,
		Snapshot:  snap,
		CreatedAt: time.Now().Unix(),
	}
	if err := svc.store.CreateSpec(sp); err != nil {
		return nil, err
	}
	return sp, nil
}

func (svc *Service) buildSnapshot(figureID string) (string, error) {
	encs, err := svc.store.ListEncodings(figureID)
	if err != nil {
		return "", err
	}
	legends, err := svc.store.ListLegends(figureID)
	if err != nil {
		return "", err
	}
	vars, err := svc.store.ListVariables(figureID)
	if err != nil {
		return "", err
	}
	axes, err := svc.store.ListAxes(figureID)
	if err != nil {
		return "", err
	}
	excs, err := svc.store.ListExceptions(figureID)
	if err != nil {
		return "", err
	}
	snap, err := spec.BuildSnapshot(figureID, encs, legends, vars, axes, excs)
	if err != nil {
		return "", err
	}
	return snap, nil
}

// SelfCheck 返回服务健康与规模指标，用于 /api/selfcheck。
func (svc *Service) SelfCheck() (map[string]interface{}, error) {
	figs, err := svc.store.ListFigures(0)
	if err != nil {
		return nil, err
	}
	open := 0
	for _, f := range figs {
		n, err := svc.store.CountOpenAmbiguities(f.ID)
		if err != nil {
			return nil, err
		}
		open += n
	}
	return map[string]interface{}{
		"ok":               true,
		"figures":          len(figs),
		"open_ambiguities": open,
		"service":          "task251-chartambiguity",
		"checker_version":  "1.0",
	}, nil
}

// Summary 返回某图形稿的派生摘要（图层/轴/变量/图例/编码/歧义计数），便于响应与自检。
func (svc *Service) Summary(figureID string) (map[string]interface{}, error) {
	layers, err := svc.store.ListLayers(figureID)
	if err != nil {
		return nil, err
	}
	axes, err := svc.store.ListAxes(figureID)
	if err != nil {
		return nil, err
	}
	vars, err := svc.store.ListVariables(figureID)
	if err != nil {
		return nil, err
	}
	legends, err := svc.store.ListLegends(figureID)
	if err != nil {
		return nil, err
	}
	encs, err := svc.store.ListEncodings(figureID)
	if err != nil {
		return nil, err
	}
	ambigs, err := svc.store.ListAmbiguities(figureID)
	if err != nil {
		return nil, err
	}
	open := 0
	for _, a := range ambigs {
		if a.IsOpen() {
			open++
		}
	}
	return map[string]interface{}{
		"layer_summary": layer.ComputeLayerSummary(layers),
		"axes":          len(axes),
		"variables":     len(vars),
		"legends":       len(legends),
		"encodings":     len(encs),
		"ambiguities":   len(ambigs),
		"open":          open,
	}, nil
}
