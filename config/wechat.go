package config

type Wechat struct {
	AppID              string `mapstructure:"app-id" json:"app-id" yaml:"app-id"`
	AppSecret          string `mapstructure:"app-secret" json:"app-secret" yaml:"app-secret"`
	AccessTokenCacheTTL string `mapstructure:"access-token-cache-ttl" json:"access-token-cache-ttl" yaml:"access-token-cache-ttl"`
}
