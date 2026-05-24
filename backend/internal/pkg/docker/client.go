package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

// LogStream wraps a Docker log stream with proper cleanup.
type LogStream struct {
	Reader *bufio.Reader
	closer io.Closer
}

// Close releases the underlying stream resources.
func (ls *LogStream) Close() error {
	if ls.closer != nil {
		return ls.closer.Close()
	}
	return nil
}

func New(host string) (*Client, error) {
	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithVersion("1.44"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, err
	}
	return &Client{cli: cli}, nil
}

func (c *Client) ListContainers(ctx context.Context) ([]types.Container, error) {
	return c.cli.ContainerList(ctx, types.ContainerListOptions{All: true})
}

func (c *Client) ContainerInspect(ctx context.Context, id string) (types.ContainerJSON, error) {
	return c.cli.ContainerInspect(ctx, id)
}

func (c *Client) ContainerStats(ctx context.Context, id string) (types.StatsJSON, error) {
	resp, err := c.cli.ContainerStats(ctx, id, false)
	if err != nil {
		return types.StatsJSON{}, err
	}
	defer resp.Body.Close()
	var stats types.StatsJSON
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return types.StatsJSON{}, err
	}
	return stats, nil
}

func (c *Client) Logs(ctx context.Context, id string, follow bool, tail string) (*LogStream, error) {
	resp, err := c.cli.ContainerLogs(ctx, id, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
		Tail:       tail,
	})
	if err != nil {
		return nil, err
	}
	return &LogStream{
		Reader: bufio.NewReader(resp),
		closer: resp,
	}, nil
}

func (c *Client) Events(ctx context.Context) (<-chan events.Message, <-chan error) {
	f := filters.NewArgs()
	ev, errs := c.cli.Events(ctx, types.EventsOptions{
		Filters: f,
	})
	return ev, errs
}
