package slogmm

import "log/slog"

var ColorMapping = map[slog.Level]string{
	slog.LevelDebug: "#63C5DA",
	slog.LevelInfo:  "#63C5DA",
	slog.LevelWarn:  "#FFA500",
	slog.LevelError: "#FF0000",
}

type message struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Channel     string       `json:"channel,omitempty"`    // INFO: в хуке канал может быть зашит, поэтому не обязательно указывать
	Username    string       `json:"username,omitempty"`   // INFO: нужно включить интеграции для переопределения имени пользователя
	IconUrl     string       `json:"icon_url,omitempty"`   // INFO: нужно включить интеграции для переопределения автарки профиля
	IconEmoji   string       `json:"icon_emoji,omitempty"` // INFO: нужно включить интеграции для переопределения автарки профиля
	Props       props        `json:"props,omitempty"`
	Priority    string       `json:"priority,omitempty"` // не разбирался, да и нужно ли
	Type        string       `json:"type,omitempty"`     // не разбирался, да и нужно ли
}

type Attachment struct {
	Fallback   string  `json:"fallback,omitempty"`
	Color      string  `json:"color,omitempty"`
	Pretext    string  `json:"pretext,omitempty"`
	Text       string  `json:"text,omitempty"`
	AuthorName string  `json:"author_name,omitempty"`
	AuthorIcon string  `json:"author_icon,omitempty"`
	AuthorLink string  `json:"author_link,omitempty"`
	Title      string  `json:"title,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	ImageUrl   string  `json:"image_url,omitempty"`
}

type Field struct {
	Short bool   `json:"short"`
	Title string `json:"title"`
	Value string `json:"value"`
}

type props struct {
	Card string `json:"card"`
}
