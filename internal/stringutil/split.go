package stringutil

import (
	"unicode"

	"github.com/upfluence/errors"
)

const zeroRune = rune(0)

var errNotValid = errors.New("Quotes did not terminate")

func Split(str string, r rune) ([]string, error) {
	var (
		parts      []string
		cur        string
		isRune     bool
		quoteAdded bool

		lastQuote = zeroRune
		lastSplit = -1
	)

	for i, c := range str {
		switch {
		case c == lastQuote:
			lastQuote = zeroRune

			if quoteAdded {
				cur += string(c)
				quoteAdded = false
			}
		case lastQuote != zeroRune:
			cur += string(c)
		case unicode.In(c, unicode.Quotation_Mark):
			isRune = false
			lastQuote = c

			if lastSplit != i-1 {
				quoteAdded = true
				cur += string(c)
			}
		case c == r:
			if 0 == i || isRune {
				continue
			}

			lastSplit = i
			isRune = true
			parts = append(parts, cur)
			cur = ""
		default:
			isRune = false
			cur += string(c)
		}
	}

	if lastQuote != zeroRune {
		return nil, errNotValid
	}

	if len(cur) != 0 {
		parts = append(parts, cur)
	}

	return parts, nil
}
