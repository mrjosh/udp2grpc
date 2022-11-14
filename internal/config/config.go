package config

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

var Map *ConfMap

type ConfMap struct {
	Mode   string         `yaml:"mode"`
	Server *ServerConfMap `yaml:"server"`
	Client *ClientConfMap `yaml:"client"`
}

type ServerConfMap struct {
	PrivateKey string            `yaml:"privatekey"`
	Listen     string            `yaml:"listen"`
	TLS        *ServerTLSConfMap `yaml:"tls"`
	Peers      []*PeerConfMap    `yaml:"peers"`
}

func (c *ServerConfMap) FindPeer(privatekey string) (*PeerConfMap, error) {
	for _, p := range c.Peers {
		if p.PrivateKey == privatekey {
			return p, nil
		}
	}
	return nil, fmt.Errorf("could not find peer")
}

type ClientConfMap struct {
	PrivateKey          string            `yaml:"privatekey"`
	Remote              string            `yaml:"remote"`
	Listen              string            `yaml:"listen,omitempty"`
	TLS                 *ServerTLSConfMap `yaml:"tls,omitempty"`
	PersistentKeepalive int64             `yaml:"persistentKeepalive,omitempty"`
}

type ServerTLSConfMap struct {
	Insecure bool   `yaml:"insecure"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type PeerConfMap struct {
	Name          string   `yaml:"name"`
	PrivateKey    string   `yaml:"privatekey"`
	Remote        string   `yaml:"remote"`
	AvailableFrom []string `yaml:"available_from"`
}

func LoadFile(filename string) (confmap *ConfMap, err error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	if err := yaml.NewDecoder(f).Decode(&Map); err != nil {
		return nil, err
	}
	return Map, nil
}
