package slogtg

import "log/slog"

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace:      "🧘 TRACE",
	slog.LevelDebug: "🪲" + slog.LevelDebug.String(),
	slog.LevelInfo:  "📖" + slog.LevelInfo.String(),
	slog.LevelWarn:  "🛑" + slog.LevelWarn.String(),
	slog.LevelError: "👺" + slog.LevelError.String(),
	LevelFatal:      "🔥 FATAL",
}
