package locales

import (
	"embed"
	"encoding/json"
	"github.com/george012/gtbox/gtbox_log"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"path/filepath"
	"sync"
)

//go:embed *.json
var locales embed.FS

var (
	once          sync.Once
	currentLocale *aLocale
)

type Language int

const (
	LanguageWithEnglish Language = iota
	LanguageWithRussia
	LanguageWithZhHK
	LanguageWithZhCN
)

func (lg Language) LanguageTag() string {
	return [...]string{"en", "ru", "zh-HK", "zh-CN"}[lg]
}

func (lg Language) String() string {
	return [...]string{"English", "Русский", "繁體中文", "简体中文"}[lg]
}

func GetLanguageFromTag(tag string) Language {
	lang := LanguageWithEnglish
	switch tag {
	case LanguageWithEnglish.LanguageTag():
		lang = LanguageWithEnglish
	case LanguageWithRussia.LanguageTag():
		lang = LanguageWithRussia
	case LanguageWithZhHK.LanguageTag():
		lang = LanguageWithZhHK
	case LanguageWithZhCN.LanguageTag():
		lang = LanguageWithZhCN
	default:
		return LanguageWithEnglish
	}
	return lang
}

type aLocale struct {
	sync.Mutex
	localizer *i18n.Localizer
	bundle    *i18n.Bundle
}

func instanceConfig() *aLocale {
	once.Do(func() {
		currentLocale = &aLocale{}
		currentLocale.initBundle()
	})
	return currentLocale
}

// 初始化和加载资源文件（仅执行一次）
func (al *aLocale) initBundle() {
	al.bundle = i18n.NewBundle(language.English)
	al.bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	entries, err := locales.ReadDir(".")
	if err != nil {
		gtbox_log.LogErrorf("Failed to read locales directory: %v", err)
	}

	// 加载嵌入的翻译文件
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) == ".json" {
			data, err := locales.ReadFile(entry.Name())
			if err != nil {
				gtbox_log.LogErrorf("Failed to read file: %v", err)
			}
			if _, err := al.bundle.ParseMessageFileBytes(data, entry.Name()); err != nil {
				gtbox_log.LogErrorf("Failed to parse message file: %v", err)
			}
		}
	}
}

// ResetLocaleLanguage 初始化本地化器
func ResetLocaleLanguage(locale string) {
	al := instanceConfig()
	al.Lock()
	defer al.Unlock()

	al.localizer = i18n.NewLocalizer(al.bundle, locale)
}

func GetLocalesMessage(messageID string) string {
	al := instanceConfig()
	al.Lock()
	defer al.Unlock()
	if al.localizer == nil {
		gtbox_log.LogErrorf("Localizer is not initialized")
		return ""
	}
	return al.localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})
}
