package service

import (
	"testing"

	"task251-chartambiguity/internal/model"
)

// publishClean 构建一个无歧义、可直接发布冻结的图形稿，返回其 ID。
func publishClean(t *testing.T, svc *Service) string {
	t.Helper()
	fig, err := svc.CreateFigure("frozen-read-only")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	payload := ImportPayload{
		Axes:      []AxisInput{{Name: "x", Variable: "time", Unit: "s", Orientation: "x"}, {Name: "y", Variable: "temp", Unit: "K", Orientation: "y"}},
		Variables: []VariableInput{{Name: "temp", Unit: "K"}},
		Legends:   []LegendInput{{Channel: "color", Label: "temperature", Token: "blue", CoversVariable: "temp"}},
	}
	if err := svc.ImportSemantics(fig.ID, payload); err != nil {
		t.Fatalf("import: %v", err)
	}
	if _, err := svc.DeclareEncoding(fig.ID, "", "temp", "color", "blue"); err != nil {
		t.Fatalf("declare encoding: %v", err)
	}
	if _, err := svc.RunCheck(fig.ID); err != nil {
		t.Fatalf("check: %v", err)
	}
	sp, err := svc.CreateSpec(fig.ID)
	if err != nil {
		t.Fatalf("create spec: %v", err)
	}
	if _, err := svc.PublishSpec(fig.ID, sp.ID); err != nil {
		t.Fatalf("publish spec: %v", err)
	}
	return fig.ID
}

// 冻结后再次复核必须被只读约束拒绝，且不得改写编码状态、歧义记录或图形稿状态。
func TestRunCheckRejectedAfterFreeze(t *testing.T) {
	svc := newTestService(t)
	fid := publishClean(t, svc)

	// 冻结前的编码语义应已固化。
	encsBefore, err := svc.Store().ListEncodings(fid)
	if err != nil {
		t.Fatalf("list encodings: %v", err)
	}
	if len(encsBefore) != 1 || encsBefore[0].Status != model.EncodingStatusValid {
		t.Fatalf("pre-freeze encoding state unexpected: %+v", encsBefore)
	}

	if _, err := svc.RunCheck(fid); err == nil {
		t.Fatal("RunCheck succeeded on a frozen figure; freeze read-only constraint missing")
	}

	// 冻结后基础语义应保持不变：编码状态、图形稿状态。
	encsAfter, err := svc.Store().ListEncodings(fid)
	if err != nil {
		t.Fatalf("list encodings after: %v", err)
	}
	if len(encsAfter) != 1 || encsAfter[0].Status != model.EncodingStatusValid {
		t.Fatalf("encoding state was mutated after freeze: %+v", encsAfter)
	}
	f, err := svc.GetFigure(fid)
	if err != nil {
		t.Fatalf("get figure after: %v", err)
	}
	if f.Status != model.FigureStatusFrozen {
		t.Fatalf("figure status was mutated after freeze: %s", f.Status)
	}
}
