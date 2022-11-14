package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

var Map *ConfMap

type ConfMap struct {
	Mode   string         `yaml:"mode"`
	Server *ServerConfMap `yaml:"server"`
}

type ServerConfMap struct {
	PrivateKey string            `yaml:"privatekey"`
	Listen     string            `yaml:"listen"`
	TLS        *ServerTLSConfMap `yaml:"tls"`
	Peers      []*PeerConfMap    `yaml:"peers"`
}

type ServerTLSConfMap struct {
	Insecure bool   `yaml:"insecure"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type PeerConfMap struct {
	Name          string   `yaml:"name"`
	PublicKey     string   `yaml:"publickey"`
	Remote        string   `yaml:"remote"`
	AvailableFrom []string `yaml:"available_from"`
}

func LoadFile(filename string) (err error) {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	return yaml.NewDecoder(f).Decode(&Map)
}
