package config

import "encoding/json"

func jsonUnmarshal(data []byte, cfg *Config) error {
	return json.Unmarshal(data, cfg)
}
