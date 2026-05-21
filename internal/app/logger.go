package app

import (
	"log/slog"
	"os"

	"github.com/golang-cz/devslog"
	slogmulti "github.com/samber/slog-multi"

	"gl.eda1.ru/go/go-service-template/config"
	"gl.eda1.ru/go/go-service-template/pkg/slogger/slogelk"
)

func slogger(cfg *config.Log) *slog.Logger {
	handlers := make([]slog.Handler, 0, 4)
	handlers = append(handlers, consoleHandler(cfg.Level))

	elk := elkHandler(cfg)
	if elk != nil {
		handlers = append(handlers, elk)
	}

	logger := slog.New(
		slogmulti.Fanout(
			handlers...,
		),
	)

	slog.SetDefault(logger)
	return logger
}

const (
	LevelTrace = slog.Level(-8)
	LevelFatal = slog.Level(12)
)

var levelNames = map[slog.Leveler]string{
	LevelTrace: "TRACE",
	LevelFatal: "FATAL",
}

func convertLevel(l string) slog.Level {
	switch l {
	case "trace":
		return LevelTrace
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "fatal":
		return LevelFatal
	default:
		return slog.LevelInfo
	}
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelLabel, exists := levelNames[level]
		if !exists {
			levelLabel = level.String()
		}

		a.Value = slog.StringValue(levelLabel)
	}

	return a
}

func consoleHandler(level string) slog.Handler {
	slogOpts := &slog.HandlerOptions{
		// AddSource:   cfg.AddSource,
		Level:       convertLevel(level),
		ReplaceAttr: replaceAttr,
	}

	opts := &devslog.Options{
		HandlerOptions:    slogOpts,
		MaxSlicePrintSize: 4,
		SortKeys:          true,
		NewLineAfterLog:   true,
		StringerFormatter: true,
	}

	return devslog.NewHandler(os.Stdout, opts)
}

func elkHandler(cfg *config.Log) slog.Handler {
	if cfg.ElkRabbitUri == "" {
		return nil
	}

	level := cfg.ElkLevel
	if level == "" {
		level = cfg.Level
	}

	return (&slogelk.Option{
		Level:    convertLevel(level),
		Url:      cfg.ElkRabbitUri,
		Partner:  cfg.Partner,
		Category: cfg.ElkCategory,
		Exchange: cfg.ElkExchange,
		// AddSource: true,
		RoutingKey:  "",
		Mandatory:   false,
		Immediate:   false,
		ReplaceAttr: replaceAttr,
	}).NewAmqpHandler()
}
