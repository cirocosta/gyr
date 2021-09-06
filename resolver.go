package gyr

// very much the same as [1], but specific to git resolution
//
// [1]: https://github.com/google/ko/blob/7f145a7e1057f2f2a099827ab9d7e66a31cd1868/pkg/resolve/resolve.go
//

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/dprotaso/go-yit"
	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v3"
)

type ReferenceResolver interface {
	// Resolve takes a "relative" reference and transforms it into an
	// absolute reference of some form.
	//
	// For instance, an image resolver might turn
	// 'gyr+container-image://library/ubuntu' into
	// 'index.docker.io/library/ubuntu@sha256:9d6a8699fb5c9c39cf08a0871bd6'
	//
	Resolve(ctx context.Context, reference string) (string, error)

	// Prefix indicate what prefix this resolver is aimed at resolving.
	//
	// for instance, a 'mergurial' resolver would advertise the prefix
	// they're interested in as 'Prefix () -> 'gyr+hg://'.
	//
	Prefix() string
}

// Resolver resolves references in yaml documents, mutating
// `gyr+*://`-annotated string fields to other values.
//
type Resolver struct {
	// ReferenceResolvers stores a set of reference resolvers that have
	// been registered to be part of the resolution.
	//
	resolvers []ReferenceResolver
}

func NewResolver(resolvers ...ReferenceResolver) *Resolver {
	return &Resolver{
		resolvers: resolvers,
	}
}

// Resolve goes through a series of YAML documents and then replaces valid
// `gyr` references to full revisions.
//
// for instance:
//
// 	---
// 	foo: gyr+gh://monero-project/monero#master
//	---
// 	bars:
// 	  - gyr+gh://monero-project/monero#master
//
// would resolve to
//
// 	---
// 	foo: 8fde011dbeb56ab92a909710567b964186671247
//	---
// 	bars:
// 	  - 8fde011dbeb56ab92a909710567b964186671247
//
// ps.: `gyr+gh://` is _not_ searched for in the middle of strings - it'll only
// match against leaf nodes prefixed with that.
//
func (r *Resolver) Resolve(ctx context.Context, docs ...*yaml.Node) error {
	refs, err := r.buildRefsMap(docs)
	if err != nil {
		return fmt.Errorf("build refs map: %w", err)
	}

	sm, err := r.resolveReferences(ctx, refs)
	if err != nil {
		return fmt.Errorf("resolve from remotes: %w", err)
	}

	err = r.updateDocuments(refs, sm)
	if err != nil {
		return fmt.Errorf("update documents: %w", err)
	}

	return nil
}

func (r *Resolver) buildRefsMap(
	docs []*yaml.Node,
) (map[string][]*yaml.Node, error) {
	refs := make(map[string][]*yaml.Node)

	for _, doc := range docs {
		it := r.refsFromDoc(doc)

		for node, ok := it(); ok; node, ok = it() {
			ref := strings.TrimSpace(node.Value)
			refs[ref] = append(refs[ref], node)
		}
	}

	return refs, nil
}

func (r *Resolver) refsFromDoc(doc *yaml.Node) yit.Iterator {
	it := yit.FromNode(doc).
		RecurseNodes().
		Filter(yit.StringValue)

	return it.Filter(func(node *yaml.Node) bool {
		return r.isSupportedScheme(node.Value)
	})
}

func (r *Resolver) isSupportedScheme(ref string) bool {
	_, found := r.resolverForRef(ref)
	return found
}

func (r *Resolver) resolverForRef(ref string) (ReferenceResolver, bool) {
	for _, resolver := range r.resolvers {
		if strings.HasPrefix(ref, resolver.Prefix()) {
			return resolver, true
		}
	}

	return nil, false
}

func (r *Resolver) resolveReferences(
	ctx context.Context, refs map[string][]*yaml.Node,
) (*sync.Map, error) {
	var (
		sm   sync.Map
		errg errgroup.Group
	)

	for ref := range refs {
		ref := ref

		errg.Go(func() error {
			resolver, _ := r.resolverForRef(ref)
			result, err := resolver.Resolve(ctx, ref)
			if err != nil {
				return fmt.Errorf("%s resolve: %w",
					resolver.Prefix(), err,
				)
			}

			sm.Store(ref, result)
			return nil
		})
	}

	if err := errg.Wait(); err != nil {
		return nil, fmt.Errorf("wait: %w", err)
	}

	return &sm, nil
}

func (r *Resolver) updateDocuments(
	refs map[string][]*yaml.Node, sm *sync.Map,
) error {
	for ref, nodes := range refs {
		revision, ok := sm.Load(ref)
		if !ok {
			return fmt.Errorf("resolve ref not found: %s", ref)
		}

		for _, node := range nodes {
			node.Value = revision.(string)
		}
	}

	return nil
}
