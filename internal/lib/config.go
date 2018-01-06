package lib

// Options contains all settings read from the configuration file
type Options struct {
	Redis struct {
		URI       string
		Password  string
		Subscribe []string
	}
	Endpoint struct {
		URI            string
		XkeyHeader     string
		SoftXkeyHeader string
		BanHeader      string
		BanURLHeader   string
	}
}
