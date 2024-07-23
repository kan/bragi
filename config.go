package main

import (
	"log"
	"os"
	"strings"

	"github.com/knadh/koanf"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
)

type Config struct {
	Port        string   `koanf:"port"`
	UseAI       bool     `koanf:"use_ai"`
	Dictironary []string `koanf:"dictionary"`
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
		"port":   "1234",
		"use_ai": true,
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