package config

type System struct {
	RouterPrefix       string `mapstructure:"router-prefix" json:"router-prefix" yaml:"router-prefix"`
	Addr               int    `mapstructure:"addr" json:"addr" yaml:"addr"` // 端口值
	Env                string `mapstructure:"env" json:"env" yaml:"env"`    // 运行环境: local / dev / production
	UseRedis           bool   `mapstructure:"use-redis" json:"use-redis" yaml:"use-redis"`
	DisableAutoMigrate bool   `mapstructure:"disable-auto-migrate" json:"disable-auto-migrate" yaml:"disable-auto-migrate"`
}
