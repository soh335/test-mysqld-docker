package mysqltest

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

type Mysqld struct {
	port      string
	host      string
	container string
}

func (m *Mysqld) DSN() string {
	return fmt.Sprintf("%s@tcp(%s:%s)/%s", "root", m.host, m.port, "test")
}

func (m *Mysqld) Stop() error {
	if err := exec.Command("docker", "kill", m.container).Run(); err != nil {
		return err
	}
	if err := exec.Command("docker", "rm", "-v", m.container).Run(); err != nil {
		return err
	}
	return nil
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
	container, err := chomp(cmd.Output())
	if err != nil {
		return nil, err
	}

	host, err := host(ctx, container)
	if err != nil {
		return nil, err
	}

	mysqld := &Mysqld{
		port:      port,
		host:      host,
		container: container,
	}
	dsn := mysqld.DSN()

	connect := time.NewTicker(time.Second)

LOOP:
	for {
		select {
		case <-ctx.Done():
			mysqld.Stop()
			return nil, ctx.Err()
		case <-connect.C:
			db, err := sql.Open("mysql", dsn)
			if err != nil {
				mysqld.Stop()
				return nil, err
			}
			if err := db.PingContext(ctx); err != nil {
				continue
			}
			break LOOP
		}
	}

	return mysqld, nil
}

func dockerRunCommand(ctx context.Context, tag string, port string) (*exec.Cmd, error) {
	var args = []string{"run"}
	if inDockerContainer() {
		cmd := exec.CommandContext(ctx, "docker", "inspect", "--format={{.HostConfig.NetworkMode}}", os.Getenv("HOSTNAME"))
		network, err := chomp(cmd.Output())
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
		host, err := chomp(cmd.Output())
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

func chomp(b []byte, err error) (string, error) {
	if err != nil {
		return "", err
	}
	return string(bytes.TrimRight(b, "\n")), nil
}
