package config

type Database struct {
	Driver string
	Source string
}

var DatabaseConfig = new(Database)
