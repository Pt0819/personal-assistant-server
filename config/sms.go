package config

type SMS struct {
	Provider string       `mapstructure:"provider" json:"provider" yaml:"provider"`
	Mock     SMSMock      `mapstructure:"mock" json:"mock" yaml:"mock"`
	Montnets SMSMontnets  `mapstructure:"montnets" json:"montnets" yaml:"montnets"`
}

type SMSMock struct {
	FixedCode string `mapstructure:"fixed-code" json:"fixed-code" yaml:"fixed-code"`
}

type SMSMontnets struct {
	APIURL   string `mapstructure:"api-url" json:"api-url" yaml:"api-url"`
	UserID   string `mapstructure:"user-id" json:"user-id" yaml:"user-id"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
	APIKey   string `mapstructure:"apikey" json:"apikey" yaml:"apikey"`
}
