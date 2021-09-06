package gyr

import (
	"context"
	"fmt"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/jenkins-x/go-scm/scm/factory"
)

type SCMResolver struct {
	client *scm.Client
}

var _ ReferenceResolver = (*SCMResolver)(nil)

func NewSCMResolver() (*SCMResolver, error) {
	client, err := factory.NewClient("", "", "")
	if err != nil {
		return nil, fmt.Errorf("new client from env: %w", err)
	}

	return &SCMResolver{client: client}, nil
}

func (r *SCMResolver) Prefix() string {
	return "gyr+gh://"
}

func (r *SCMResolver) Resolve(
	ctx context.Context, str string,
) (string, error) {
	repository, reference, err := r.repositoryAndReference(str)
	if err != nil {
		return "", err
	}

	commit, _, err := r.client.Git.FindCommit(ctx, repository, reference)
	if err != nil {
		return "", fmt.Errorf(
			"find commit for repo=%s ref=%s: %w", repository, reference, err,
		)
	}

	return commit.Sha, nil
}

func (r *SCMResolver) repositoryAndReference(
	str string,
) (string, string, error) {
	parts := strings.Split(str[len(r.Prefix()):], "#")
	if len(parts) != 2 {
		return "", "", fmt.Errorf(
			"expected 2 parts, got %d", len(parts),
		)
	}

	repository, reference := parts[0], parts[1]
	return repository, reference, nil
}
