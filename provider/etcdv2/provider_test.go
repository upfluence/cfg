package etcdv2

import (
	"context"
	"os"
	"testing"

	"github.com/coreos/etcd/client"
	"github.com/stretchr/testify/assert"
	"github.com/upfluence/pkg/testutil"

	"github.com/upfluence/cfg/provider/providertest"
)

func TestNewProvider(t *testing.T) {
	for _, tt := range []struct {
		name  string
		opts  []Option
		errfn testutil.ErrorAssertion
	}{
		{
			name:  "no options work",
			errfn: testutil.NoError(),
		},
		{
			name:  "no endpoints",
			opts:  []Option{SetEndpoints()},
			errfn: testutil.ErrorEqual(client.ErrNoEndpoints),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.opts...)

			assert.Equal(t, "etcd", p.StructTag())
			if tag := p.StructTag(); tag != "etcd" {
				t.Errorf(`p.StructTag() = %q [ want: "etcd" ]`, tag)
			}

			errfn, ok := p.(interface{ Err() error })

			if !ok {
				errfn = noErrBuilder{}
			}

			tt.errfn(t, errfn.Err())
		})
	}
}

type noErrBuilder struct{}

func (noErrBuilder) Err() error { return nil }

func setKey(k, v string) func(*testing.T, client.KeysAPI) {
	return func(t *testing.T, kapi client.KeysAPI) {
		_, err := kapi.Set(context.Background(), k, v, &client.SetOptions{})

		assert.Nil(t, err)
	}
}

func deleteKey(k string, dir bool) func(*testing.T, client.KeysAPI) {
	return func(t *testing.T, kapi client.KeysAPI) {
		opts := client.DeleteOptions{}

		if dir {
			opts.Dir = true
			opts.Recursive = true
		}

		_, err := kapi.Delete(context.Background(), k, &opts)

		assert.Nil(t, err)
	}
}

func TestProvide(t *testing.T) {
	url := os.Getenv("ETCD_URL")

	if url == "" {
		t.Skip("ETCD_URL env var is not provided skipping tests")
	}

	for _, tt := range []struct {
		name           string
		prefix         string
		setup, cleanup func(*testing.T, client.KeysAPI)

		ks []string

		pfn providertest.ProvideAssertion
	}{
		{
			name:    "directory not found",
			prefix:  "/foobar",
			ks:      []string{"fiz"},
			setup:   func(*testing.T, client.KeysAPI) {},
			cleanup: func(*testing.T, client.KeysAPI) {},
			pfn: providertest.AssertError(func(t testing.TB, err error) {
				etcderr, ok := err.(client.Error)

				assert.True(t, ok)
				assert.Equal(t, client.ErrorCodeKeyNotFound, etcderr.Code)
			}),
		},
		{
			name:    "prefix is a key",
			prefix:  "/foobar",
			ks:      []string{"fiz"},
			setup:   setKey("/foobar", "buz"),
			cleanup: deleteKey("/foobar", false),
			pfn:     providertest.AssertError(testutil.ErrorEqual(errNotDirectory)),
		},
		{
			name:    "prefix is a  flat dir",
			prefix:  "/foobar",
			ks:      []string{"fiz", "fuz"},
			setup:   setKey("/foobar/fuz", "buz"),
			cleanup: deleteKey("/foobar", true),
			pfn: providertest.AssertValues(
				providertest.MapValue(map[string]string{"fuz": "buz"}),
			),
		},
		{
			name:    "prefix is a nested dir",
			prefix:  "/foobar",
			ks:      []string{"fuz.fiz", "fuz"},
			setup:   setKey("/foobar/fuz/biz", "buz"),
			cleanup: deleteKey("/foobar", true),
			pfn: providertest.AssertValues(
				providertest.MapValue(map[string]string{"fuz.biz": "buz"}),
			),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(SetEndpoints(url), WithPrefix(tt.prefix))

			ep, ok := p.(*Provider)

			assert.True(t, ok)

			tt.setup(t, ep.KeysAPI)
			defer tt.cleanup(t, ep.KeysAPI)

			for _, k := range tt.ks {
				tt.pfn(t, p, k)
			}
		})
	}
}
