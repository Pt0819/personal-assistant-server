package config

type Security struct {
	EncryptionKey string `mapstructure:"encryption-key" json:"encryption-key" yaml:"encryption-key"`
	MaxDevices    int    `mapstructure:"max-devices" json:"max-devices" yaml:"max-devices"`
}
