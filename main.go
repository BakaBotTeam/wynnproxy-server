package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	mcproxy "wynnproxyserver/proxy"
)

var configName = "wynnproxy-config.json"

type Config struct {
	LocalHost  string `json:"localhost"`
	LocalPort  int    `json:"localport"`
	RemoteHost string `json:"remotehost"`
	RemotePort int    `json:"remoteport"`
	MOTD       string `json:"motd"`
}

func LoadConfig() (*Config, error) {
	file, err := os.ReadFile(configName)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(file, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveConfig(config *Config) error {
	data, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(configName, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	config, err := LoadConfig()

	if err != nil {
		config = &Config{
			LocalHost:  "127.0.0.1",
			LocalPort:  25565,
			RemoteHost: "play.wynncraft.com",
			RemotePort: 25565,
			MOTD:       "WynnCraft-Proxy 🩷From BakaTeam",
		}
		err = SaveConfig(config)
		if err != nil {
			log.Println("Failed to save config file", err)
		}
	}

	localAddress := config.LocalHost + ":" + strconv.Itoa(config.LocalPort)
	serverAddress := config.RemoteHost + ":" + strconv.Itoa(config.RemotePort)
	motd := config.MOTD
	server := mcproxy.MinecraftProxyServer{
		Listen: localAddress,
		Remote: serverAddress,
		MOTD:   motd,
	}
	log.Println("Proxy server started")
	server.StartServer()
}
