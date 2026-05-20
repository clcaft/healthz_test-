package slogtg

import (
	"fmt"
	"log/slog"
	"strings"

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

	level, ok := levelNames[record.Level]
	if !ok {
		level = record.Level.String()
	}

	// handler formatter
	message := fmt.Sprintf("%s: %s\n------------\n\n", level, record.Message)
	message += attrToTelegramMessage("", attrs)
	return message
}

func attrToTelegramMessage(base string, attrs []slog.Attr) string {
	var message strings.Builder

	for i := range attrs {
		attr := attrs[i]
		k := base + attr.Key
		v := attr.Value
		kind := attr.Value.Kind()

		if kind == slog.KindGroup {
			message.WriteString(attrToTelegramMessage(k+".", v.Group()))
		} else {
			message.WriteString(fmt.Sprintf("%s: %s\n", k, slogcommon.ValueToString(v)))
		}
	}

	return message.String()
}
