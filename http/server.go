package http

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	"wynnproxyserver/proxy"
	"wynnproxyserver/utils"
)

type HttpServer struct {
	ListenPort int
	Secret     string
}

type Result struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

// var proxies = make(map[string]proxy.Proxy)

func (server HttpServer) InitServer() {
	http.HandleFunc("/verify", func(writer http.ResponseWriter, request *http.Request) {
		query := request.URL.Query()
		name := query.Get("name")
		timestamp := query.Get("ts")
		host := query.Get("host")
		port := query.Get("port")
		sign := query.Get("sign")

		ts, err := strconv.ParseInt(timestamp, 10, 64)

		if err != nil {
			bytes, _ := json.Marshal(Result{Code: 103, Message: "Failed to parse timestamp"})
			fmt.Fprint(writer, string(bytes))
			return
		}

		_, err = strconv.ParseInt(port, 10, 64)

		if err != nil {
			bytes, _ := json.Marshal(Result{Code: 104, Message: "Failed to parse port"})
			fmt.Fprint(writer, string(bytes))
		}

		if time.Now().UnixMilli()-ts >= 1000*100 {
			bytes, _ := json.Marshal(Result{Code: 101, Message: "Timestamp expired"})
			fmt.Fprint(writer, string(bytes))
			return
		}

		if !strings.HasSuffix(host, ".proxy.wynncraft.com") {
			bytes, _ := json.Marshal(Result{Code: 102, Message: "Invalid host"})
			fmt.Fprint(writer, string(bytes))
			return
		}

		signRaw := fmt.Sprintf("%s:%s:%s:%s:%s", name, timestamp, host, port, server.Secret)
		expectedSign := utils.SHA256Hash(signRaw)
		if expectedSign == sign {
			proxy.Users[name] = proxy.User{
				ServerAddress: host + ":" + port,
			}

			bytes, _ := json.Marshal(Result{Code: 0, Message: "OK"})
			fmt.Fprint(writer, string(bytes))
		} else {
			bytes, _ := json.Marshal(Result{Code: 100, Message: "Invalid signature"})
			fmt.Fprint(writer, string(bytes))
		}
	})

	log.Println("Http server is running on localhost:" + strconv.Itoa(server.ListenPort))

	err := http.ListenAndServe(":"+strconv.Itoa(server.ListenPort), nil)
	if err != nil {
		log.Println("Failed to start HttpServer:", err)
	}
}
