package dflt

import "context"

type Provider struct{}

func (Provider) StructTag() string { return "default" }

func (Provider) DefaultFieldValue(_ string) string { return "" }

func (Provider) JoinFieldKeys(_, key string) string {
	return key
}

func (Provider) SubKeys(context.Context, string) ([]string, error) {
	return nil, nil
}

func (Provider) Provide(_ context.Context, k string) (string, bool, error) {
	if k == "" {
		return "", false, nil
	}

	return k, true, nil
}
