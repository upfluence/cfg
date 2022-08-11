package etcdv2

import (
	"context"
	"strings"
	"sync"

	etcd "github.com/coreos/etcd/client"
	"github.com/upfluence/errors"

	"github.com/upfluence/cfg/provider"
)

var (
	errNotDirectory = errors.New("Key given is not a directory")

	defaultOptions = options{
		config: etcd.Config{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	}
)

type options struct {
	config etcd.Config
	prefix string
}

type Option func(*options)

func SetEndpoints(es ...string) func(*options) {
	return func(opts *options) {
		opts.config.Endpoints = es
	}
}

func WithPrefix(p string) func(*options) {
	return func(opts *options) {
		opts.prefix = p
	}
}

func AddEndpoints(es ...string) func(*options) {
	return func(opts *options) {
		opts.config.Endpoints = append(opts.config.Endpoints, es...)
	}
}

type Provider struct {
	sync.Once
	p provider.Provider

	etcd.KeysAPI
	prefix string

	kfn provider.KeyFn
}

func NewProvider(opts ...Option) provider.Provider {
	var o = defaultOptions

	for _, opt := range opts {
		opt(&o)
	}

	c, err := etcd.New(o.config)

	if err != nil {
		return provider.ProvideError("etcd", err)
	}

	suffixedPrefixed := o.prefix

	if !strings.HasSuffix(o.prefix, "/") {
		suffixedPrefixed += "/"
	}

	return &Provider{
		KeysAPI: etcd.NewKeysAPI(c),
		prefix:  o.prefix,
		kfn: func(s string) string {
			return strings.NewReplacer("/", ".").Replace(
				strings.TrimPrefix(s, suffixedPrefixed),
			)
		},
	}
}

func (p *Provider) StructTag() string { return "etcd" }

func (p *Provider) buildProvider(ctx context.Context) {
	resp, err := p.Get(
		ctx,
		p.prefix,
		&etcd.GetOptions{Recursive: true, Sort: true},
	)

	if err != nil {
		p.p = provider.ProvideError("etcd", err)
		return
	}

	if !resp.Node.Dir {
		p.p = provider.ProvideError("etcd", errNotDirectory)
		return
	}

	res := make(map[string]string)

	p.parseNode(resp.Node, res)

	p.p = provider.NewStaticProvider(
		"etcd",
		res,
		func(s string) string { return s },
	)
}

func (p *Provider) parseNode(n *etcd.Node, vs map[string]string) {
	if !n.Dir {
		vs[p.kfn(n.Key)] = n.Value
		return
	}

	for _, n := range n.Nodes {
		p.parseNode(n, vs)
	}
}

func (p *Provider) Provide(ctx context.Context, k string) (string, bool, error) {
	p.Do(func() { p.buildProvider(ctx) })

	return p.p.Provide(ctx, k)
}
