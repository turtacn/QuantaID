package ldap

import (
	"context"

	ber "github.com/go-asn1-ber/asn1-ber"
	"go.uber.org/zap"
)

type Entry struct {
	DN         string
	Attributes map[string][]string
}

func (s *Server) HandleSearch(ctx context.Context, messageID int64, req *ber.Packet) *ber.Packet {
	// SearchRequest ::= [APPLICATION 3] SEQUENCE {
	//     baseObject      LDAPDN,
	//     scope           ENUMERATED { baseObject (0), singleLevel (1), wholeSubtree (2) },
	//     derefAliases    ENUMERATED { neverDerefAliases (0), ... },
	//     sizeLimit       INTEGER (0 .. maxInt),
	//     timeLimit       INTEGER (0 .. maxInt),
	//     typesOnly       BOOLEAN,
	//     filter          Filter,
	//     attributes      AttributeSelection }

	if len(req.Children) < 8 {
		return encodeLDAPResult(messageID, ApplicationSearchResultDone, LDAPResultProtocolError, "", "Invalid SearchRequest")
	}

	baseObject := req.Children[0].Value.(string)
	scope := req.Children[1].Value.(int64)
	// derefAliases := req.Children[2].Value.(int64)
	// sizeLimit := req.Children[3].Value.(int64)
	// timeLimit := req.Children[4].Value.(int64)
	// typesOnly := req.Children[5].Value.(bool)
	filterPacket := req.Children[6]
	attributesPacket := req.Children[7]

	requestedAttrs := []string{}
	for _, child := range attributesPacket.Children {
		requestedAttrs = append(requestedAttrs, child.Value.(string))
	}

	s.logger.Debug("Search Request", zap.String("base", baseObject), zap.Int64("scope", scope))

	entries, err := s.virtualTree.Search(ctx, baseObject, int(scope), filterPacket, requestedAttrs)
	if err != nil {
		s.logger.Error("Search error", zap.Error(err))
		return encodeLDAPResult(messageID, ApplicationSearchResultDone, LDAPResultOperationsError, "", err.Error())
	}

	var responses []*ber.Packet
	for _, entry := range entries {
		responses = append(responses, encodeSearchResultEntry(messageID, entry))
	}

	responses = append(responses, encodeLDAPResult(messageID, ApplicationSearchResultDone, LDAPResultSuccess, "", ""))

	// Hack: We need to return multiple packets, but the interface returns one.
	// We will return a special sequence that handleConnection will need to unwrap,
	// OR we modify handleConnection to accept []*ber.Packet.
	// Let's modify handleConnection to handle this specific case or change the return type.
	// For now, to avoid changing signature in all files immediately, let's return a "Sequence" packet containing all responses,
	// and handleConnection can iterate if it's a special type? No that's hacky.
	// Best way: Change Handle* to return []*ber.Packet.

	// Since I can't change signature easily in one diff without breaking others temporarily,
	// I will pack them into a container packet and unwrap in handleConnection.
	container := ber.Encode(ber.ClassContext, ber.TypeConstructed, 0, nil, "MultiResponse")
	for _, r := range responses {
		container.AppendChild(r)
	}
	return container
}

// Helper to encode Entry
func encodeSearchResultEntry(messageID int64, entry *Entry) *ber.Packet {
	// LDAP Message Sequence
	packet := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAP Message")
	packet.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, messageID, "MessageID"))

	// SearchResultEntry ProtocolOp
	op := ber.Encode(ber.ClassApplication, ber.TypeConstructed, ApplicationSearchResultEntry, nil, "SearchResultEntry")
	op.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, entry.DN, "objectName"))

	attrs := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "attributes")
	for k, vals := range entry.Attributes {
		seq := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "PartialAttribute")
		seq.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, k, "type"))
		set := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "vals")
		for _, v := range vals {
			set.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v, "value"))
		}
		seq.AppendChild(set)
		attrs.AppendChild(seq)
	}
	op.AppendChild(attrs)

	packet.AppendChild(op)
	return packet
}
