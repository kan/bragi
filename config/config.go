package config

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/pkg/errors"
)

type Config struct {
	Port           string   `koanf:"port" json:"port"`
	AdminPort      string   `koanf:"admin_port" json:"admin_port"`
	UseAI          bool     `koanf:"use_ai" json:"use_ai"`
	UseLisp        bool     `koanf:"use_lisp" json:"use_lisp"`
	YearFormat     string   `koanf:"year_format" json:"year_format"`
	MonthFormat    string   `koanf:"month_format" json:"month_format"`
	DateFormat     string   `koanf:"date_format" json:"date_format"`
	DateTimeFormat string   `koanf:"date_time_format" json:"date_time_format"`
	TimeZone       string   `koanf:"time_zone" json:"time_zone"`
	Dictionary     []string `koanf:"dictionary" json:"dictionary"`
	DictPath       string   `koanf:"dict_path" json:"dict_path"`
}

func (config *Config) GetCacheDir() (string, error) {
	dir := config.DictPath
	if dir == "" {
		cdir, err := os.UserCacheDir()
		if err != nil {
			return "", errors.WithStack(err)
		}
		dir = filepath.Join(cdir, "bragi", "dict")
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", errors.WithStack(err)
	}

	return dir, nil
}

func LoadConfig(filename string) (*Config, error) {
	k := koanf.New(".")

	// 設定ファイル(TOML)から読み込み
	if _, err := os.Stat(filename); err == nil {
		k.Load(file.Provider(filename), toml.Parser())
	}

	// 環境変数から読み込み
	k.Load(env.ProviderWithValue("BRG_", ".", func(s, v string) (string, interface{}) {
		key := strings.ToLower(strings.TrimPrefix(s, "BRG_"))
		log.Printf("%s => %s: %s\n", s, key, v)
		if key == "dictionary" {
			return key, strings.Split(v, ",")
		}

		return key, v
	}), nil)

	// 初期値を設定
	defaults := map[string]interface{}{
		"port":             "1234",
		"admin_port":       "8080",
		"use_ai":           true,
		"use_lisp":         true,
		"year_format":      "2006年",
		"month_format":     "2006年1月",
		"date_format":      "2006年1月2日",
		"date_time_format": "2006年1月2日 15時4分",
		"time_zone":        "Asia/Tokyo",
	}
	for key, val := range defaults {
		if !k.Exists(key) {
			k.Set(key, val)
		}
	}

	var config Config
	err := k.Unmarshal("", &config)

	return &config, err
}
