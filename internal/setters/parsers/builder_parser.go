package parsers

type Parser interface {
	Parse(string, bool) (interface{}, error)
}
