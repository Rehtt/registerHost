package main

import (
	"context"
	"encoding/json"
	"errors"
	"iter"
	"log/slog"
	"net"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types/container"
)

func dockerInspect(ctx context.Context, id string) ([]*container.InspectResponse, error) {
	var data []*container.InspectResponse
	out, err := exec.CommandContext(ctx, "docker", "inspect", id).CombinedOutput()
	if err != nil {
		return nil, errors.New(string(out))
	}
	err = json.Unmarshal(out, &data)
	return data, err
}

func getDockerNameIp(ctx context.Context, id string) (string, map[string]net.IP) {
	dockerInfo, err := dockerInspect(ctx, id)
	if err != nil {
		slog.Error("docker inspect failed", "id", id, "err", err)
		return "", nil
	}
	for _, v := range dockerInfo {
		name := strings.TrimPrefix(v.Name, "/")
		if name != id {
			continue
		}
		ips := make(map[string]net.IP, len(v.NetworkSettings.Networks))
		for networkName, network := range v.NetworkSettings.Networks {
			ips[networkName] = net.ParseIP(network.IPAddress)
		}
		return name, ips
	}
	return "", nil
}

func parseDockerComposeContainerName(dockerCompose string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for v := range strings.SplitSeq(dockerCompose, "\n") {
			v := strings.ToLower(v)
			if i := strings.Index(v, "container_name:"); i > 0 {
				if m := strings.Index(v, "#"); m > 0 && m < i {
					continue
				}
				if s := strings.Split(v, ":"); len(s) == 2 {
					yield(strings.TrimSpace(s[1]))
				}
			}
		}
	}
}
