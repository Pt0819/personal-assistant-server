package config

type Server struct {
	JWT      JWT      `mapstructure:"jwt" json:"jwt" yaml:"jwt"`
	Zap      Zap      `mapstructure:"zap" json:"zap" yaml:"zap"`
	Redis    Redis    `mapstructure:"redis" json:"redis" yaml:"redis"`
	System   System   `mapstructure:"system" json:"system" yaml:"system"`
	Mysql    Mysql    `mapstructure:"mysql" json:"mysql" yaml:"mysql"`
	Wechat   Wechat   `mapstructure:"wechat" json:"wechat" yaml:"wechat"`
	Email    Email    `mapstructure:"email" json:"email" yaml:"email"`
	SMS      SMS      `mapstructure:"sms" json:"sms" yaml:"sms"`
	Grpc     Grpc     `mapstructure:"grpc" json:"grpc" yaml:"grpc"`
	Oss      Oss      `mapstructure:"oss" json:"oss" yaml:"oss"`
	Security Security `mapstructure:"security" json:"security" yaml:"security"`
}
