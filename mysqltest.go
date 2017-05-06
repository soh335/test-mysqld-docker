package mysqltest

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

type Mysqld struct {
	port string
	host string
}

func (m *Mysqld) DSN() string {
	return fmt.Sprintf("%s@tcp(%s:%s)/%s", "root", m.host, m.port, "test")
}

func NewMysqld(ctx context.Context, tag string) (*Mysqld, error) {
	port, err := port()
	if err != nil {
		return nil, err
	}

	cmd, err := dockerRunCommand(ctx, tag, port)
	if err != nil {
		return nil, err
	}
	container, err := chomp(cmd.Output)
	if err != nil {
		return nil, err
	}

	host, err := host(ctx, container)
	if err != nil {
		return nil, err
	}

	mysqld := &Mysqld{
		port: port,
		host: host,
	}
	dsn := mysqld.DSN()

	connect := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			_ = exec.Command("docker", "kill", container).Run()
			_ = exec.Command("docker", "rm", "-v", container).Run()
			return nil, ctx.Err()
		case <-connect.C:
			db, err := sql.Open("mysql", dsn)
			if err != nil {
				return nil, err
			}
			if err := db.PingContext(ctx); err != nil {
				continue
			}
			break
		}
	}

	return mysqld, nil
}

func dockerRunCommand(ctx context.Context, tag string, port string) (*exec.Cmd, error) {
	var args = []string{"run"}
	if inDockerContainer() {
		cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{.HostConfig.NetworkMode}}", os.Getenv("HOSTNAME"))
		network, err := chomp(cmd.Output)
		if err != nil {
			return nil, err
		}
		args = append(args, "--network", string(network))
	} else {
		args = append(args, "-p", fmt.Sprintf("%s:3306", port))
	}
	args = append(args, "-e", "MYSQL_ALLOW_EMPTY_PASSWORD=1")
	args = append(args, "-e", "MYSQL_DATABASE=test")
	args = append(args, "-d", tag)
	return exec.CommandContext(ctx, "docker", args...), nil
}

func port() (string, error) {
	if inDockerContainer() {
		return "3306", nil
	} else {
		return emptyPort()
	}
}

func host(ctx context.Context, container string) (string, error) {
	if inDockerContainer() {
		cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", container)
		host, err := chomp(cmd.Output)
		return host, err
	} else {
		return "127.0.0.1", nil
	}
}

func inDockerContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func emptyPort() (string, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	_, port, err := net.SplitHostPort(l.Addr().String())
	if err != nil {
		return "", err
	}
	l.Close()
	return port, nil
}

type outputer func() ([]byte, error)

func chomp(o outputer) (string, error) {
	b, err := o()
	if err != nil {
		return "", err
	}
	return string(bytes.TrimRight(b, "\n")), nil
}
