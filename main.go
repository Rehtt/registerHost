package main

import (
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	goweb "github.com/Rehtt/Kit/web"
)

var (
	lock       sync.Mutex
	listenAddr = flag.String("listen", "127.0.0.1:8080", "listen addr")
)

func main() {
	flag.Parse()

	g := goweb.New()
	g.GET("/set/host/#host", func(ctx *goweb.Context) {
		host := ctx.GetUrlPathParam("host")
		remoteAddr := ctx.Request.RemoteAddr
		ip := strings.Split(remoteAddr, ":")[0]
		setSystemHost(host, net.ParseIP(ip))
	})

	g.GET("/set/docker/#dockerId/#network", func(ctx *goweb.Context) {
		dockerId := ctx.GetUrlPathParam("dockerId")
		network := ctx.GetUrlPathParam("network")
		dockerName, dockerIp := getDockerNameIp(ctx, dockerId, network)
		setSystemHost(dockerName, dockerIp)
	})

	if err := http.ListenAndServe(*listenAddr, g); err != nil {
		slog.Error("listen failed", "err", err)
	}
}

func setSystemHost(host string, ip net.IP) {
	if ip == nil {
		return
	}
	lock.Lock()
	defer lock.Unlock()

	hosts, err := os.ReadFile("/etc/hosts")
	if err != nil {
		slog.Error("set host failed", "err", err)
		return
	}
	data := strings.Split(string(hosts), "\n")
	var find bool
	for i := range data {
		v := strings.ReplaceAll(data[i], "\t", " ")
		v = strings.TrimSpace(v)
		if strings.HasSuffix(v, " "+host) {
			data[i] = ip.String() + " " + host
			find = true
			break
		}
	}

	if !find {
		data = append(data, ip.String()+" "+host)
	}

	info, _ := os.Stat("/etc/hosts")
	if err = os.WriteFile("/etc/hosts", []byte(strings.Join(data, "\n")), info.Mode()); err != nil {
		slog.Error("set host failed", "err", err)
	}
}
