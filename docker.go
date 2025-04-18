package main

import (
	"context"
	"log/slog"
	"net"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func dockerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return container.InspectResponse{}, err
	}
	defer cli.Close()
	return cli.ContainerInspect(ctx, id)
}

func getDockerNameIp(ctx context.Context, id string, network string) (string, net.IP) {
	dockerInfo, err := dockerInspect(ctx, id)
	if err != nil {
		slog.Error("docker inspect failed", "id", id, "err", err)
		return "", nil
	}
	name := strings.TrimPrefix(dockerInfo.Name, "/")
	n, ok := dockerInfo.NetworkSettings.Networks[network]
	if !ok {
		return name, nil
	}
	return name, net.IP(n.IPAddress)
}
