package util

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver            string        `mapstructure:"DB_DRIVER"`
	DBSource            string        `mapstructure:"DB_SOURCE"`
	ServerAddress       string        `mapstructure:"SERVER_ADDRESS"`
	TokenSymmetricKey   string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
}

//LoadCOnfig read configuration from file or env,
func LocalConfig(path string) (config Config, err error) {
	// path는 위치만 적고 이름은 안 적어도된다. ./app.env면 다음과 같다.
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv() // 자동으로 env에서 일치하는 값이 있는 지 확인한다.

	err = viper.ReadInConfig() // config 파일을 읽는다.
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
