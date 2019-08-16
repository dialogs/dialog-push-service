package conversion

type Config struct {
	AllowAlerts bool   `mapstructure:"allow-alerts"`
	Sound       string `mapstructure:"sound"`
	Topic       string `mapstructure:"topic"`
}
