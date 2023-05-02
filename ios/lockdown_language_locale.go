package ios

import log "github.com/sirupsen/logrus"

// LanguageConfiguration is a simple struct encapsulating a language and locale string
type LanguageConfiguration struct {
	Language           string
	Locale             string
	SupportedLocales   []string
	SupportedLanguages []string
}

const languageDomain = "com.apple.international"

// SetLanguage creates a new lockdown session for the device and sets a new language and locale.
// Changes will only be made when the value is not an empty string. To change only the locale, set language to ""
// and vice versa. If both are empty, nothing is changed.
// NOTE: Changing a language is an async operation that takes a long time. Springboard will be restarted automatically by the device.
// If you need to wait for this happen use notificationproxy.WaitUntilSpringboardStarted().
func SetLanguage(device DeviceEntry, config LanguageConfiguration) error {
	if config.Locale == "" && config.Language == "" {
		log.Debug("SetLanguage called with empty config, no changes made")
		return nil
	}
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return err
	}
	defer lockDownConn.Close()
	if config.Locale != "" {
		log.Debugf("Setting locale: %s", config.Locale)
		err := lockDownConn.SetValueForDomain("Locale", languageDomain, config.Locale)
		if err != nil {
			return err
		}
	}
	if config.Language != "" {
		log.Debugf("Setting language: %s", config.Language)
		return lockDownConn.SetValueForDomain("Language", languageDomain, config.Language)
	}
	return nil
}

// GetLanguage creates a new lockdown session for the device and retrieves the current language and locale as well as
// a list of all supported locales and languages.
// It returns a LanguageConfiguration or an error.
func GetLanguage(device DeviceEntry) (LanguageConfiguration, error) {
	lockDownConn, err := ConnectLockdownWithSession(device)
	if err != nil {
		return LanguageConfiguration{}, err
	}
	defer lockDownConn.Close()
	languageResp, err := lockDownConn.GetValueForDomain("Language", languageDomain)
	if err != nil {
		return LanguageConfiguration{}, err
	}
	localeResp, err := lockDownConn.GetValueForDomain("Locale", languageDomain)
	if err != nil {
		return LanguageConfiguration{}, err
	}

	supportedLocalesResp, err := lockDownConn.GetValueForDomain("SupportedLocales", languageDomain)
	if err != nil {
		return LanguageConfiguration{}, err
	}

	supportedLanguagesResp, err := lockDownConn.GetValueForDomain("SupportedLanguages", languageDomain)
	if err != nil {
		return LanguageConfiguration{}, err
	}
	supportedLocales := InterfaceToStringSlice(supportedLocalesResp)
	supportedLanguages := InterfaceToStringSlice(supportedLanguagesResp)
	return LanguageConfiguration{Language: languageResp.(string), Locale: localeResp.(string), SupportedLocales: supportedLocales, SupportedLanguages: supportedLanguages}, nil
}
