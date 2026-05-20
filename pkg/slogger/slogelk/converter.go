package slogelk

import (
	"encoding/json"
	"log/slog"

	slogcommon "github.com/samber/slog-common"
)

var SourceKey = "source"

type Converter func(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) string

func DefaultConverter(addSource bool, replaceAttr func(groups []string, a slog.Attr) slog.Attr, loggerAttr []slog.Attr, groups []string, record *slog.Record) string {
	// aggregate all attributes
	attrs := slogcommon.AppendRecordAttrsToAttrs(loggerAttr, groups, record)

	// developer formatters
	attrs = slogcommon.ReplaceAttrs(replaceAttr, []string{}, attrs...)
	attrs = slogcommon.RemoveEmptyAttrs(attrs)

	// formatter
	msg := attrToAqpmMessage(attrs)

	if len(msg) == 0 {
		return record.Message
	} else {
		msg["message"] = record.Message
	}

	b, _ := json.Marshal(msg)

	return string(b)
}

func attrToAqpmMessage(attrs []slog.Attr) map[string]any {
	m := make(map[string]any)

	for i := range attrs {
		attr := attrs[i]
		k := attr.Key
		v := attr.Value
		kind := attr.Value.Kind()

		if kind == slog.KindGroup {
			m[k] = attrToAqpmMessage(v.Group())
		} else {
			m[k] = slogcommon.ValueToString(v)
		}
	}

	return m
}
