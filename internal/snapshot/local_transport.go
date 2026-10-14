package snapshot

import (
	"fmt"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/storer"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/server"
)

// go-git's default file transport shells out to git-upload-pack. Install the
// in-process implementation during package initialization, before any worker
// starts. The library protocol registry is never mutated at runtime.
func init() { client.InstallProtocol("file", server.NewClient(localLoader{})) }

type localLoader struct{}

func (localLoader) Load(endpoint *transport.Endpoint) (storer.Storer, error) {
	if endpoint.Host != "" && endpoint.Host != "localhost" {
		return nil, fmt.Errorf("non-local file endpoint is unsupported")
	}
	repo, err := git.PlainOpen(endpoint.Path)
	if err != nil {
		return nil, err
	}
	return repo.Storer, nil
}
