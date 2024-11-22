package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	mcproxy "wynnproxyserver/proxy"
)

var configName = "wynnproxy-config.json"

type Config struct {
	ProxyServerInfo struct {
		ListenPort int    `json:"listenport"`
		RemoteHost string `json:"remotehost"`
		RemotePort int    `json:"remoteport"`
		MOTD       string `json:"motd"`
	} `json:"proxyserverinfo"`

	HttpServerInfo struct {
		ListenPort int    `json:"listenport"`
		Secret     string `json:"secret"`
	} `json:"httpserverinfo"`
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
		config = &Config{}
		config.ProxyServerInfo.ListenPort = 25565
		config.ProxyServerInfo.RemoteHost = "play.wynncraft.com"
		config.ProxyServerInfo.RemotePort = 25565
		config.ProxyServerInfo.MOTD = "WynnCraft-Proxy 🩷From BakaTeam"
		config.HttpServerInfo.ListenPort = 1337
		config.HttpServerInfo.Secret = "Just8Bit"
		err = SaveConfig(config)
		if err != nil {
			log.Println("Failed to save config file:", err)
		}
	}

	proxyServerInfo := config.ProxyServerInfo
	httpServerInfo := config.HttpServerInfo

	proxyPort := proxyServerInfo.ListenPort
	httpPort := httpServerInfo.ListenPort

	localAddress := "127.0.0.1:" + strconv.Itoa(proxyPort)
	serverAddress := proxyServerInfo.RemoteHost + ":" + strconv.Itoa(proxyServerInfo.RemotePort)
	motd := proxyServerInfo.MOTD

	server := mcproxy.MinecraftProxyServer{
		Listen: localAddress,
		Remote: serverAddress,
		MOTD:   motd,
	}

	log.Println("Proxy server is running on localhost:" + strconv.Itoa(proxyPort))

	go func() {
		server.StartServer()
	}()

	http.HandleFunc("/verify", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintf(writer, "暂时先这样 等客户端mod吧!")
	})

	err = http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)

	if err != nil {
		log.Println("Failed to start HttpServer:", err)
	}
}
