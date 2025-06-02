package reference

import (
	"context"
	"errors"
	"testing"
)

import (
	"phite.io/polygenic-risk-calculator/internal/model"
)

type mockBigQueryClient struct {
	stats *model.ReferenceStats
	err   error
}

func (m *mockBigQueryClient) GetReferenceStats(ctx context.Context, ancestry, trait, model string) (*model.ReferenceStats, error) {
	return m.stats, m.err
}
func (m *mockBigQueryClient) Close() error { return nil }

func TestReferenceStatsLoader(t *testing.T) {
	ctx := context.Background()

	t.Run("returns valid stats", func(t *testing.T) {
		want := &model.ReferenceStats{Mean: 1.2, Std: 0.5, Min: -2, Max: 2, Ancestry: "EUR", Trait: "height", Model: "v1"}
		client := &mockBigQueryClient{stats: want}
		got, err := client.GetReferenceStats(ctx, "EUR", "height", "v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil || got.Mean != want.Mean {
			t.Errorf("unexpected stats: %+v", got)
		}
	})

	t.Run("returns nil if no match", func(t *testing.T) {
		client := &mockBigQueryClient{stats: nil}
		got, err := client.GetReferenceStats(ctx, "AFR", "height", "v1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != nil {
			t.Errorf("expected nil for missing stats, got %+v", got)
		}
	})

	t.Run("returns error if query fails", func(t *testing.T) {
		client := &mockBigQueryClient{err: errors.New("query failed")}
		_, err := client.GetReferenceStats(ctx, "EUR", "height", "v1")
		if err == nil {
			t.Errorf("expected error for failed query")
		}
	})
}
