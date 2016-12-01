package mysqltest

import (
	"bytes"
	"database/sql"
	"fmt"
	"net"
	"os"
	"os/exec"
	"time"
)

type MysqldConfig struct {
	Tag string
}

func NewMysqldConfig() *MysqldConfig {
	return &MysqldConfig{
		Tag: "mysql:latest",
	}
}

type Mysqld struct {
	port      string
	host      string
	config    *MysqldConfig
	container string
}

func NewMysqld(config *MysqldConfig) (*Mysqld, error) {
	if config == nil {
		config = NewMysqldConfig()
	}
	mysqld := &Mysqld{
		config: config,
	}
	if err := mysqld.start(); err != nil {
		return nil, err
	}
	return mysqld, nil
}

func (m *Mysqld) DSN() string {
	return fmt.Sprintf("%s@tcp(%s:%s)/%s", "root", m.host, m.port, "test")
}

func (m *Mysqld) Stop() {
	killCointainer(m.container)
	removeContainer(m.container)
}

func (m *Mysqld) start() error {
	cmd, port, err := m.dockerRunCommand()
	if err != nil {
		return err
	}
	_container, err := cmd.Output()
	container := string(chomp(_container))
	if err != nil {
		return err
	}

	m.container = container
	m.port = port

	if inDockerContainer() {
		o, err := exec.Command("docker", "inspect", "--format={{.NetworkSettings.IPAddress}}", m.container).Output()
		if err != nil {
			return err
		}
		m.host = string(chomp(o))
	} else {
		m.host = "127.0.0.1"
	}

	timeout := time.NewTimer(time.Second * 30)
	connect := time.NewTicker(time.Second)

	for {
		select {
		case <-timeout.C:
			killCointainer(container)
			removeContainer(container)
			return fmt.Errorf("timeout: failed to connect mysqld")
		case <-connect.C:
			dsn := m.DSN()
			db, err := sql.Open("mysql", dsn)
			if err != nil {
				return err
			}
			if err := db.Ping(); err != nil {
				continue
			}
			return nil
		}
	}
}

func (m *Mysqld) dockerRunCommand() (*exec.Cmd, string, error) {
	var args = []string{"run"}
	var port string
	if inDockerContainer() {
		o, err := exec.Command("docker", "inspect", "--format={{.HostConfig.NetworkMode}}", os.Getenv("HOSTNAME")).Output()
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
	args = append(args, "-d", m.config.Tag)
	return exec.Command("docker", args...), port, nil
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
