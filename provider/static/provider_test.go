package static

import "testing"

func TestNewProvider(t *testing.T) {
	for _, tt := range []struct {
		name string
		in   interface{}
		out  string
	}{
		{name: "correct value", in: map[string]string{"foo": "bar"}, out: "json"},
		{
			name: "return error provider",
			in:   map[float64]string{2: "bar"},
			out:  "static",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.in)
			if tag := p.StructTag(); tag != tt.out {
				t.Errorf("p.StructTag() = %v [ want %v ]", tag, tt.out)
			}
		})
	}
}
