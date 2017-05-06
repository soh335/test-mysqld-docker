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

func NewMysqld(ctx context.Context, tag string) (*Mysqld, error) {
	mysqld := &Mysqld{}
	if err := mysqld.start(ctx, tag); err != nil {
		return nil, err
	}
	return mysqld, nil
}

func (m *Mysqld) DSN() string {
	return fmt.Sprintf("%s@tcp(%s:%s)/%s", "root", m.host, m.port, "test")
}

func (m *Mysqld) start(ctx context.Context, tag string) error {
	cmd, port, err := m.dockerRunCommand(ctx, tag)
	if err != nil {
		return err
	}
	_container, err := cmd.Output()
	container := string(chomp(_container))
	if err != nil {
		return err
	}

	m.port = port

	if inDockerContainer() {
		o, err := exec.CommandContext(ctx, "docker", "inspect", "--format={{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", container).Output()
		if err != nil {
			return err
		}
		m.host = string(chomp(o))
	} else {
		m.host = "127.0.0.1"
	}

	connect := time.NewTicker(time.Second)
	dsn := m.DSN()

	for {
		select {
		case <-ctx.Done():
			killCointainer(container)
			removeContainer(container)
			return ctx.Err()
		case <-connect.C:
			log.Println("ping...", dsn)
			db, err := sql.Open("mysql", dsn)
			if err != nil {
				return err
			}
			if err := db.PingContext(ctx); err != nil {
				continue
			}
			return nil
		}
	}
}

func (m *Mysqld) dockerRunCommand(ctx context.Context, tag string) (*exec.Cmd, string, error) {
	var args = []string{"run"}
	var port string
	if inDockerContainer() {
		o, err := exec.CommandContext(ctx, "docker", "inspect", "--format={{.HostConfig.NetworkMode}}", os.Getenv("HOSTNAME")).Output()
		if err != nil {
			return nil, "", err
		}
		network := chomp(o)
		args = append(args, "--network", string(network))
		port = "3306"
	} else {
		forward, err := emptyPort()
		if err != nil {
			return nil, "", err
		}
		args = append(args, "-p", fmt.Sprintf("%s:3306", forward))
		port = forward
	}
	args = append(args, "-e", "MYSQL_ALLOW_EMPTY_PASSWORD=1")
	args = append(args, "-e", "MYSQL_DATABASE=test")
	args = append(args, "-d", tag)
	return exec.CommandContext(ctx, "docker", args...), port, nil
}

func inDockerContainer() bool {
	_, err := os.Stat("/.dockerenv")
	return err == nil
}

func killCointainer(id string) error {
	return exec.Command("docker", "kill", id).Run()
}

func removeContainer(id string) error {
	return exec.Command("docker", "rm", "-v", id).Run()
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

func chomp(v []byte) []byte {
	return bytes.TrimRight(v, "\n")
}
