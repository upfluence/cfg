package parsers

type Parser interface {
	parse(string, bool) (interface{}, error)
}
