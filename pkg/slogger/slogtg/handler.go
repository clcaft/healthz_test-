package slogtg

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	slogcommon "github.com/samber/slog-common"
)

// curl -X POST \
//      -H 'Content-Type: application/json' \
//      -d '{"chat_id": "<your-chat-id>", "text": "This is a test from curl", "disable_notification": true}' \
//      https://api.telegram.org/bot<your-bot-token>/sendMessage

type Option struct {
	// log level (default: Fatal)
	Level slog.Leveler

	// Telegram bot token
	Token string

	ChatID              int64
	MessageThreadID     int64
	DisableNotification bool

	// optional: customize Telegram message builder
	Converter Converter
	// MessageConfigurator MessageConfigurator

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o *Option) NewTelegramHandler() slog.Handler {
	if o.Level == nil {
		o.Level = LevelFatal
	}

	if o.Token == "" {
		panic("missing Telegram token")
	}

	if o.ChatID == 0 {
		panic("missing Telegram chat id")
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	return &telegramHandler{
		option: o,
		client: http.DefaultClient,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*telegramHandler)(nil)

type telegramHandler struct {
	option *Option
	client *http.Client
	attrs  []slog.Attr
	groups []string
}

func (h *telegramHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

type message struct {
	ChatID              int64  `json:"chat_id,string"`
	MessageThreadID     int64  `json:"message_thread_id,omitempty"`
	Text                string `json:"text"`
	DisableNotification bool   `json:"disable_notification"`
}

func (h *telegramHandler) message(ctx context.Context, msg string) error {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(message{
		ChatID:              h.option.ChatID,
		MessageThreadID:     h.option.MessageThreadID,
		Text:                msg,
		DisableNotification: h.option.DisableNotification,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.telegram.org/bot"+h.option.Token+"/sendMessage", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

//nolint:gocritic // implement slog.Handler interface
func (h *telegramHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.message(ctx, h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, h.attrs, h.groups, &record))
}

func (h *telegramHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &telegramHandler{
		option: h.option,
		client: h.client,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *telegramHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &telegramHandler{
		option: h.option,
		client: h.client,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
