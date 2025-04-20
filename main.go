package main

import (
	"context"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/Rehtt/Kit/slice"
	goweb "github.com/Rehtt/Kit/web"
)

var (
	lock              sync.Mutex
	listenAddr        = flag.String("listen", "127.0.0.1:8080", "listen addr")
	dockercomposeFile slice.FlagSetArray[string]
	network           = flag.String("network", "", "docker network")
	suffix            = flag.String("suffix", ".docker.local", "suffix")
)

func main() {
	flag.Var(&dockercomposeFile, "compose", "docker compose")
	flag.Parse()

	if len(dockercomposeFile.Get()) > 0 {
		for _, v := range dockercomposeFile.Get() {
			if *network == "" {
				slog.Error("network is empty")
				os.Exit(1)
			}
			data, err := os.ReadFile(v)
			if err != nil {
				slog.Error("read docker compose failed", "err", err)
				os.Exit(1)
			}
			for name := range parseDockerComposeContainerName(string(data)) {
				dockerName, dockerIps := getDockerNameIp(context.Background(), name)
				setSystemHost(dockerName, dockerIps[*network])
			}
		}
		return
	}

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
		dockerName, dockerIp := getDockerNameIp(ctx, dockerId)
		setSystemHost(dockerName, dockerIp[network])
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

	host += *suffix

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
