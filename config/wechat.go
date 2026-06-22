package config

type Wechat struct {
	AppID              string `mapstructure:"app-id" json:"app-id" yaml:"app-id"`
	AppSecret          string `mapstructure:"app-secret" json:"app-secret" yaml:"app-secret"`
	AccessTokenCacheTTL    string `mapstructure:"access-token-cache-ttl" json:"access-token-cache-ttl" yaml:"access-token-cache-ttl"`
	OpenPlatformAppID      string `mapstructure:"open-platform-app-id" json:"open-platform-app-id" yaml:"open-platform-app-id"`
	OpenPlatformAppSecret  string `mapstructure:"open-platform-app-secret" json:"open-platform-app-secret" yaml:"open-platform-app-secret"`
	WebRedirectURI         string `mapstructure:"web-redirect-uri" json:"web-redirect-uri" yaml:"web-redirect-uri"`
}
