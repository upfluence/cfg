package cfg

import "context"

type Provider interface {
	StructTag() string
	Provide(context.Context, string) (string, bool, error)
}
