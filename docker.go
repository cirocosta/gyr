package gyr

import (
	"context"
	"fmt"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

type DockerResolver struct {
	opts []remote.Option
}

var _ ReferenceResolver = (*DockerResolver)(nil)

func NewDockerResolver() (*DockerResolver, error) {
	return &DockerResolver{opts: []remote.Option{
		remote.WithAuth(authn.Anonymous),
	}}, nil
}

func (r *DockerResolver) Prefix() string {
	return "gyr+docker://"
}

func (r *DockerResolver) Resolve(
	ctx context.Context, str string,
) (string, error) {
	imageName := str[len(r.Prefix()):]

	imageReference, err := name.ParseReference(imageName)
	if err != nil {
		return "", fmt.Errorf("parse reference '%s': %w",
			imageName, err,
		)
	}

	img, err := remote.Image(imageReference, r.opts...)
	if err != nil {
		return "", fmt.Errorf("image '%s': %w",
			imageReference.String(), err,
		)
	}

	digest, err := img.Digest()
	if err != nil {
		return "", fmt.Errorf("img digest: %w", err)
	}

	return imageReference.Name() + "@" + digest.String(), nil
}
