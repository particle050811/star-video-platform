package relation

import (
	"testing"
	relationrepo "video-platform/biz/repository/relation"
	userrepo "video-platform/biz/repository/user"
)

func TestBuildSocialList(t *testing.T) {
	users := []userrepo.UserProfile{
		{
			ID:        1,
			Username:  "alice",
			AvatarURL: "/static/avatars/alice.png",
		},
		{
			ID:        2,
			Username:  "bob",
			AvatarURL: "/static/avatars/bob.png",
		},
	}

	got := buildSocialList(&relationrepo.RelationListResult{
		Users:      users,
		Total:      99,
		NextCursor: 2,
		HasMore:    true,
	})
	if got.Total != 99 {
		t.Fatalf("expected total %d, got %d", 99, got.Total)
	}
	if len(got.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got.Items))
	}
	if got.NextCursor != "2" || !got.HasMore {
		t.Fatalf("unexpected cursor response: %+v", got)
	}
	if got.Items[0].Id != "1" {
		t.Fatalf("expected first item id %q, got %q", "1", got.Items[0].Id)
	}
	if got.Items[0].Username != "alice" {
		t.Fatalf("expected first item username %q, got %q", "alice", got.Items[0].Username)
	}
	if got.Items[0].AvatarUrl != "/static/avatars/alice.png" {
		t.Fatalf("expected first item avatar %q, got %q", "/static/avatars/alice.png", got.Items[0].AvatarUrl)
	}
	if got.Items[1].Id != "2" {
		t.Fatalf("expected second item id %q, got %q", "2", got.Items[1].Id)
	}
	if got.Items[1].Username != "bob" {
		t.Fatalf("expected second item username %q, got %q", "bob", got.Items[1].Username)
	}
	if got.Items[1].AvatarUrl != "/static/avatars/bob.png" {
		t.Fatalf("expected second item avatar %q, got %q", "/static/avatars/bob.png", got.Items[1].AvatarUrl)
	}
}

func TestBuildSocialListReturnsEmptyItems(t *testing.T) {
	got := buildSocialList(&relationrepo.RelationListResult{})
	if got == nil {
		t.Fatal("expected non-nil data")
	}
	if got.Items == nil {
		t.Fatal("expected non-nil empty items")
	}
	if len(got.Items) != 0 || got.Total != 0 || got.NextCursor != "" || got.HasMore {
		t.Fatalf("unexpected empty result: %+v", got)
	}
}

func TestBuildSocialListHandlesNilResult(t *testing.T) {
	got := buildSocialList(nil)
	if got == nil {
		t.Fatal("expected non-nil data")
	}
	if got.Items == nil {
		t.Fatal("expected non-nil empty items")
	}
	if len(got.Items) != 0 {
		t.Fatalf("expected empty items, got %+v", got.Items)
	}
}
