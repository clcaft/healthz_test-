package slogmm

import (
	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

const SourceKey = "source"

var propsCardKey = "PropsCard"

type Converter func(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) *message

func DefaultConverter(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) *message {
	// aggregate all attributes
	attrs := slogcommon.AppendRecordAttrsToAttrs(loggerAttr, groups, record)

	// developer formatters
	if addSource {
		attrs = append(attrs, slogcommon.Source(SourceKey, record))
	}
	attrs = slogcommon.ReplaceAttrs(replaceAttr, []string{}, attrs...)
	attrs = slogcommon.RemoveEmptyAttrs(attrs)

	// handler formatter
	message := &message{}
	message.Text = record.Message
	message.Attachments = []Attachment{
		{
			Color:  ColorMapping[record.Level],
			Fields: []Field{},
		},
	}

	attrToMattermostMessage("", attrs, message)
	return message
}

func attrToMattermostMessage(base string, attrs []slog.Attr, message *message) {
	for i := range attrs {
		attr := attrs[i]
		k := attr.Key
		v := attr.Value
		kind := attr.Value.Kind()

		switch {
		case kind == slog.KindGroup:
			attrToMattermostMessage(base+k+".", v.Group(), message)
		case k == propsCardKey:
			message.Props = props{Card: slogcommon.ValueToString(v)}
		default:
			field := Field{}
			field.Title = base + k
			field.Value = slogcommon.ValueToString(v)
			message.Attachments[0].Fields = append(message.Attachments[0].Fields, field)
		}
	}
}
