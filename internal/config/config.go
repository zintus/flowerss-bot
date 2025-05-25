package config

import (
	"bytes" // Added
	"fmt"
	"text/template"

	"github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	tb "gopkg.in/telebot.v3"

	"github.com/zintus/flowerss-bot/internal/i18n"
)

type RunType string

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"

	ProjectName          string = "flowerss"
	BotToken             string
	Socks5               string
	TelegraphToken       []string
	TelegraphAccountName string
	TelegraphAuthorName  string = "flowerss-bot"
	TelegraphAuthorURL   string

	// EnableTelegraph 是否启用telegraph
	EnableTelegraph       bool = false
	PreviewText           int  = 0
	DisableWebPagePreview bool = false
	mysqlConfig           *mysql.Config
	SQLitePath            string
	EnableMysql           bool = false

	// UpdateInterval rss抓取间隔
	UpdateInterval int = 10

	// ErrorThreshold rss源抓取错误阈值
	ErrorThreshold uint = 100

	// MessageTpl rss更新推送模版
	MessageTpl *template.Template

	// MessageMode telegram消息渲染模式
	MessageMode tb.ParseMode

	// TelegramEndpoint telegram bot 服务器地址，默认为空
	TelegramEndpoint string = tb.DefaultApiURL

	// UserAgent User-Agent
	UserAgent string

	// RunMode 运行模式 Release / Debug
	RunMode RunType = ReleaseMode

	// AllowUsers 允许使用bot的用户
	AllowUsers []int64

	// DBLogMode 是否打印数据库日志
	DBLogMode bool = false
)

const (
	defaultMessageTplMode = tb.ModeHTML
	defaultMessageTpl     = `<b>{{.SourceTitle}}</b>{{ if .PreviewText }}
{{ .L "feed_update_preview_header" }}
{{.PreviewText}}
-----------------------------
{{- end}}{{if .EnableTelegraph}}
{{.ContentTitle}} <a href="{{.TelegraphURL}}">{{ .L "feed_update_telegraph_link_text" }}</a> | <a href="{{.RawLink}}">{{ .L "feed_update_original_link_text" }}</a>
{{- else }}
<a href="{{.RawLink}}">{{.ContentTitle}}</a>
{{- end }}
{{.Tags}}
`
	defaultMessageMarkdownTpl = `** {{.SourceTitle}} **{{ if .PreviewText }}
{{ .L "feed_update_preview_header" }}
{{.PreviewText}}
-----------------------------
{{- end}}{{if .EnableTelegraph}}
{{.ContentTitle}} [{{ .L "feed_update_telegraph_link_text" }}]({{.TelegraphURL}}) | [{{ .L "feed_update_original_link_text" }}]({{.RawLink}})
{{- else }}
[{{.ContentTitle}}]({{.RawLink}})
{{- end }}
{{.Tags}}
`
	TestMode    RunType = "Test"
	ReleaseMode RunType = "Release"
)

type TplData struct {
	SourceTitle     string
	ContentTitle    string
	RawLink         string
	PreviewText     string
	TelegraphURL    string
	Tags            string
	EnableTelegraph bool
	LangCode        string // Added for localization
}

// AppVersionInfo returns a localized string with version, commit and date.
// It now accepts a langCode for localization.
func AppVersionInfo(langCode string) string {
	// Ensure i18n.Localize is available and i18n system is initialized.
	// Default to "en" if langCode is empty or invalid, Localize should handle this.
	format := i18n.Localize(langCode, "version_info_format")
	return fmt.Sprintf(format, version, commit, date)
}

func (td *TplData) Render(mode tb.ParseMode) (string, error) {
	var tpl string
	if mode == tb.ModeMarkdown || mode == tb.ModeMarkdownV2 {
		tpl = defaultMessageMarkdownTpl
	} else {
		tpl = defaultMessageTpl
	}

	funcMap := template.FuncMap{
		"L": func(key string, args ...interface{}) string {
			langToUse := td.LangCode
			if langToUse == "" {
				langToUse = "en" // Fallback to "en"
			}
			return i18n.Localize(langToUse, key, args...)
		},
	}

	var buf bytes.Buffer
	t, err := template.New("message").Funcs(funcMap).Parse(tpl)
	if err != nil {
		return "", err
	}

	if err := t.Execute(&buf, td); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// GetString get string config value by key
func GetString(key string) string {
	var value string
	if viper.IsSet(key) {
		value = viper.GetString(key)
	}

	return value
}

func GetMysqlDSN() string {
	return mysqlConfig.FormatDSN()
}
