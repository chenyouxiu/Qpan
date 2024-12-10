package models

type Config struct {
	Mysql  Mysql  `yaml:"mysql"`
	Server Server `yaml:"server"`
}
type Mysql struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type Server struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}
