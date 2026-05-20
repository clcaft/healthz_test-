package slogmm

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	slogcommon "github.com/samber/slog-common"
)

type Option struct {
	// log level (default: Warn)
	Level slog.Leveler

	// Mattermost webhook
	WebhookURL   string
	Channel      string
	Username     string
	IconUrl      string
	IconEmoji    string
	PropsCardKey string // default: PropsCard

	// optional: customize Mattermost message builder
	Converter Converter

	// optional: see slog.HandlerOptions
	AddSource   bool
	ReplaceAttr func(groups []string, a slog.Attr) slog.Attr
}

func (o *Option) NewMattermostHandler() slog.Handler {
	if o.Level == nil {
		o.Level = slog.LevelWarn
	}

	if o.WebhookURL == "" {
		panic("missing Mattermost hook")
	}

	if o.Converter == nil {
		o.Converter = DefaultConverter
	}

	if o.PropsCardKey != "" {
		propsCardKey = o.PropsCardKey
	}

	return &mattermostHandler{
		option: o,
		client: http.DefaultClient,
		attrs:  []slog.Attr{},
		groups: []string{},
	}
}

var _ slog.Handler = (*mattermostHandler)(nil)

type mattermostHandler struct {
	option *Option
	client *http.Client
	attrs  []slog.Attr
	groups []string
}

func (h *mattermostHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.option.Level.Level()
}

func (h *mattermostHandler) message(ctx context.Context, msg *message) error {
	var buf bytes.Buffer
	json.NewEncoder(&buf).Encode(msg)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.option.WebhookURL, &buf)
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
func (h *mattermostHandler) Handle(ctx context.Context, record slog.Record) error {
	message := h.option.Converter(h.option.AddSource, h.option.ReplaceAttr, h.attrs, h.groups, &record)

	message.Channel = h.option.Channel
	message.Username = h.option.Username
	message.IconUrl = h.option.IconUrl
	message.IconEmoji = h.option.IconEmoji

	return h.message(ctx, message)
}

func (h *mattermostHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &mattermostHandler{
		option: h.option,
		client: h.client,
		attrs:  slogcommon.AppendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups: h.groups,
	}
}

func (h *mattermostHandler) WithGroup(name string) slog.Handler {
	// https://cs.opensource.google/go/x/exp/+/46b07846:slog/handler.go;l=247
	if name == "" {
		return h
	}

	return &mattermostHandler{
		option: h.option,
		client: h.client,
		attrs:  h.attrs,
		groups: append(h.groups, name),
	}
}
