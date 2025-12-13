package ldap

import (
	"strings"

	ber "github.com/go-asn1-ber/asn1-ber"
)

func MatchesFilter(entry *Entry, filter *ber.Packet) bool {
	switch filter.Tag {
	case FilterAnd:
		for _, child := range filter.Children {
			if !MatchesFilter(entry, child) {
				return false
			}
		}
		return true
	case FilterOr:
		for _, child := range filter.Children {
			if MatchesFilter(entry, child) {
				return true
			}
		}
		return false
	case FilterNot:
		return !MatchesFilter(entry, filter.Children[0])
	case FilterEqualityMatch:
		return checkAttributeMatch(entry, filter, func(val, filterVal string) bool {
			return strings.EqualFold(val, filterVal)
		})
	case FilterSubstrings:
		// Sequence { type, substrings }
		attrType := filter.Children[0].Value.(string)
		// substrings sequence
		seq := filter.Children[1]

		vals, ok := entry.Attributes[attrType]
		if !ok {
			return false
		}

		for _, v := range vals {
			// Check if v matches substring pattern
			// Simplification: check only if "any" or "initial" match
			// Real impl needs to handle sequence of initial, any, final
			matched := true
			for _, sub := range seq.Children {
				subVal := sub.Value.(string)
				// sub.Tag: 0=initial, 1=any, 2=final
				switch sub.Tag {
				case 0:
					if !strings.HasPrefix(strings.ToLower(v), strings.ToLower(subVal)) {
						matched = false
					}
				case 1:
					if !strings.Contains(strings.ToLower(v), strings.ToLower(subVal)) {
						matched = false
					}
				case 2:
					if !strings.HasSuffix(strings.ToLower(v), strings.ToLower(subVal)) {
						matched = false
					}
				}
			}
			if matched {
				return true
			}
		}
		return false

	case FilterPresent:
		attrType := filter.Data.String()
		_, ok := entry.Attributes[attrType]
		return ok
	default:
		// Unsupported filter type, default to true or false?
		// False is safer
		return false
	}
}

func checkAttributeMatch(entry *Entry, filter *ber.Packet, matcher func(string, string) bool) bool {
	// AttributeValueAssertion ::= SEQUENCE {
	//     attributeDesc   AttributeDescription,
	//     assertionValue  AssertionValue }
	if len(filter.Children) < 2 {
		return false
	}
	attr := filter.Children[0].Value.(string)
	val := filter.Children[1].Value.(string)

	entryVals, ok := entry.Attributes[attr]
	if !ok {
		// objectClass check is special?
		// No, objectClass is just an attribute
		return false
	}

	for _, v := range entryVals {
		if matcher(v, val) {
			return true
		}
	}
	return false
}
