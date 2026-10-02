package i18n

const DefaultLocale = "id-ID"

var supportedLocales = map[string]struct{}{
	DefaultLocale: {},
}

func Normalize(locale string) string {
	if _, ok := supportedLocales[locale]; ok {
		return locale
	}
	return DefaultLocale
}

func Supported() []string {
	return []string{DefaultLocale}
}
