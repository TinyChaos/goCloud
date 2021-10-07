package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const URL = "localhost:80"
const VisitLog = "./tmp/log/visit.log"
const (
	SUCCESS = 200
	FAIL    = 500
)

type LogInfo struct {
	Time 	   string
	IP         string
	StatusCode int
	Output     string
}

func main() {
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(URL, nil))
}

func GetGolangVersion(env string) string {
	var version string
	if version = os.Getenv(env); version == "" {
		version = "none" //若没有则返回none
	}
	return version
}

func handler(w http.ResponseWriter, r *http.Request) {
	code := FAIL
	// 当访问 localhost/healthz 时，应返回 200
	if r.URL.Path == "/healthz" {
		code = SUCCESS
		version := GetGolangVersion("VERSION")
		w.Header().Set("version", version) // 读取当前系统的环境变量中的 VERSION 配置，并写入 response header
	}
	w.WriteHeader(code)
	for k,v := range r.Header {
		for _,value := range v {
			w.Header().Set(k,value) //接收客户端 request，并将 request 中带的 header 写入 response header
		}
	}
	w.Write([]byte("Hello,world!"))
	RecordLog(r,code) //Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
}

func GetIP(r *http.Request) string {
	xForwardedFor := r.Header.Get(`X-Forwarded-For`)
	ip := strings.TrimSpace(strings.Split(xForwardedFor, `,`)[0])
	if ip == "" {
		ip = strings.TrimSpace(r.Header.Get(`X-Real-Ip`))
		if ip == "" {
			_ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr))
			if err != nil {
				panic(err)
			} else {
				ip = _ip
			}
		}
	}
	return ip
}

func init() {
	errorLogFile := VisitLog
	f , err := os.OpenFile(errorLogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func WriteLog(logPath string,content LogInfo) {
	f , err := os.OpenFile(logPath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0644)
	defer f.Close()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	jsonBytes,err := json.Marshal(content)
	if err != nil {
		return
	}
	_,_ = f.Write(jsonBytes)
}

func RecordLog(r *http.Request,code int){
	var info LogInfo
	info.Time = time.Now().String()
	info.IP = GetIP(r)
	info.StatusCode = code
	bodyLen := r.ContentLength
	if bodyLen != 0 {
		body := make([]byte,bodyLen)
		_, err := r.Body.Read(body)
		if err != nil {
			return
		}
		info.Output = string(body)
	}else{
		info.Output = ""
	}
	info.Output += `\n`
	WriteLog(VisitLog, info)
}
