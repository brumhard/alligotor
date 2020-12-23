package test

type APIConfig struct {
	Port     int    `config:"env=PORT,flag=p"`
	LogLevel string `config:"file=loglevel"`
}

type DBConfig struct {
	Password string
	LogLevel string `config:"file=loglevel"`
}
