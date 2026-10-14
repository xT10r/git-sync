package snapshot

import (
	"context"
	"path/filepath"
	"testing"

	"git-sync/internal/testutil"
)

func BenchmarkUnchangedSnapshot(b *testing.B) {
	remote := testutil.NewRemote(b)
	remote.Commit("config", "v1")
	p, err := New(Config{URL: remote.Path, Branch: "master", Root: filepath.Join(b.TempDir(), "root")})
	if err != nil {
		b.Fatal(err)
	}
	defer p.Close()
	if _, err = p.Sync(context.Background()); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, err := p.Sync(context.Background())
		if err != nil || r.Published {
			b.Fatalf("unexpected no-op: %+v %v", r, err)
		}
	}
}
