package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type (
	Postgres struct {
		Host        string `yaml:"host"`
		Port        int64  `yaml:"port"`
		User        string `yaml:"user"`
		Pass        string `yaml:"pass"`
		Dbname      string `yaml:"dbname"`
		Connections *struct {
			MaxIdle     int64 `yaml:"max_idle"`
			MaxOpen     int64 `yaml:"max_open"`
			MaxLifetime int64 `yaml:"max_lifetime"`
		} `yaml:"connections"`
	}

	Config struct {
		Postgres *Postgres `yaml:"postgres"`
	}
)

func NewConfig(path string) (config Config, err error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("could not read file: %w", err)
	}

	if err = yaml.Unmarshal(bytes, &config); err != nil {
		return config, fmt.Errorf("could not unmarshal config: %w", err)
	}

	return config, nil
}

// Connect raw sql connect
func (p *Postgres) Connect() (db *sql.DB, err error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host,
		p.Port,
		p.User,
		p.Pass,
		p.Dbname,
	)

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("could not open db: %w", err)
	}

	_, err = db.Exec("SELECT 1")
	if err != nil {
		return nil, fmt.Errorf("could not test db connection: %w", err)
	}

	return db, nil
}
