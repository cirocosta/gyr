package gyr_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/stretchr/testify/require"

	"github.com/cirocosta/gyr"
)

type deterministicResolver struct {
	prefix  string
	mapping map[string]string
}

func (r *deterministicResolver) Resolve(ctx context.Context, ref string) (string, error) {
	v, found := r.mapping[ref]
	if !found {
		return "", fmt.Errorf("not found")
	}

	return v, nil
}

func (r *deterministicResolver) Prefix() string { return r.prefix }

func TestResolver(t *testing.T) {
	for _, tc := range []struct {
		name      string
		input     string
		expected  string
		resolvers []gyr.ReferenceResolver
		err       string
	}{
		{
			name:      "no resolver, no change",
			resolvers: nil,
			input:     `foo: gyr+foo://test`,
			expected: heredoc.Doc(`---
			foo: gyr+foo://test
			`),
		},

		{
			name: "no resolver for scheme, no change",
			resolvers: []gyr.ReferenceResolver{
				&deterministicResolver{
					prefix: "gyr+blabla://",
				},
			},
			input: `foo: gyr+foo://test`,
			expected: heredoc.Doc(`---
			foo: gyr+foo://test
			`),
		},

		{
			name: "scheme registered, failing resolution",
			resolvers: []gyr.ReferenceResolver{
				&deterministicResolver{
					prefix: "gyr+foo://",
					mapping: map[string]string{
						"gyr+foo://nononon": "x",
					},
				},
			},
			input: `foo: gyr+foo://test`,
			err:   "resolve: not found",
		},

		{
			name: "scheme registered, working resolution",
			resolvers: []gyr.ReferenceResolver{
				&deterministicResolver{
					prefix: "gyr+foo://",
					mapping: map[string]string{
						"gyr+foo://test": "bar",
					},
				},
			},
			input: `foo: gyr+foo://test`,
			expected: heredoc.Doc(`---
			foo: bar
			`),
		},

		{
			name: "multiple schemes registered, working resolution",
			resolvers: []gyr.ReferenceResolver{
				&deterministicResolver{
					prefix: "gyr+first://",
					mapping: map[string]string{
						"gyr+first://first": "1",
					},
				},
				&deterministicResolver{
					prefix: "gyr+second://",
					mapping: map[string]string{
						"gyr+second://second": "2",
					},
				},
			},
			input: heredoc.Doc(`---
			foo: gyr+first://first
			caz: gyr+second://second
			`),
			expected: heredoc.Doc(`---
			foo: "1"
			caz: "2"
			`),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nodes, err := gyr.NodesFromReader(strings.NewReader(tc.input))
			require.NoError(t, err)

			err = gyr.NewResolver(tc.resolvers...).
				Resolve(context.Background(), nodes...)
			if tc.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.err)
				return
			}

			require.NoError(t, err)

			var buf bytes.Buffer
			err = gyr.WriteYAML(&buf, nodes)
			require.NoError(t, err)

			require.Equal(t, tc.expected, buf.String())
		})
	}
}
