package service

import (
	"testing"

	"task251-chartambiguity/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	st, err := store.Open(":memory:")
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return New(st)
}

func TestCreateSpecRequiresExistingPublishableFigure(t *testing.T) {
	svc := newTestService(t)
	figure, err := svc.CreateFigure("spec gate")
	if err != nil {
		t.Fatalf("create figure: %v", err)
	}
	if _, err := svc.CreateSpec(figure.ID); err == nil {
		t.Fatal("spec created before figure was checked")
	}
}
