package config

type Email struct {
	Provider string    `mapstructure:"provider" json:"provider" yaml:"provider"`
	From     string    `mapstructure:"from" json:"from" yaml:"from"`
	SMTP     EmailSMTP `mapstructure:"smtp" json:"smtp" yaml:"smtp"`
}

type EmailSMTP struct {
	Host     string `mapstructure:"host" json:"host" yaml:"host"`
	Port     int    `mapstructure:"port" json:"port" yaml:"port"`
	Username string `mapstructure:"username" json:"username" yaml:"username"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}
