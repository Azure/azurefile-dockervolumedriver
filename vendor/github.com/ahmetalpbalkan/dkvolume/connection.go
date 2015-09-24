package dkvolume

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/listenbuffer"
	"github.com/docker/libcontainer/user"
)

// Code extracted from Docker.

// See for more details:
// https://github.com/docker/docker/blob/master/api/server/unix_socket.go
// https://github.com/docker/docker/blob/master/api/server/tcp_socket.go

// TLSConfig is the structure that represents the TLS configuration.
type TLSConfig struct {
	CA          string
	Certificate string
	Key         string
	Verify      bool
}

func newUnixSocket(path, group string, activate <-chan struct{}) (net.Listener, error) {
	if err := syscall.Unlink(path); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	mask := syscall.Umask(0777)
	defer syscall.Umask(mask)
	l, err := listenbuffer.NewListenBuffer("unix", path, activate)
	if err != nil {
		return nil, err
	}
	if err := setSocketGroup(path, group); err != nil {
		l.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0660); err != nil {
		l.Close()
		return nil, err
	}
	return l, nil
}

func newTCPSocket(addr string, config *TLSConfig, activate <-chan struct{}) (net.Listener, error) {
	l, err := listenbuffer.NewListenBuffer("tcp", addr, activate)
	if err != nil {
		return nil, err
	}
	if config != nil {
		if l, err = setupTLS(l, config); err != nil {
			return nil, err
		}
	}
	return l, nil
}

func setupTLS(l net.Listener, config *TLSConfig) (net.Listener, error) {
	tlsCert, err := tls.LoadX509KeyPair(config.Certificate, config.Key)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("Could not load X509 key pair (%s, %s): %v", config.Certificate, config.Key, err)
		}
		return nil, fmt.Errorf("Error reading X509 key pair (%s, %s): %q. Make sure the key is encrypted.",
			config.Certificate, config.Key, err)
	}
	tlsConfig := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: []tls.Certificate{tlsCert},
		// Avoid fallback on insecure SSL protocols
		MinVersion: tls.VersionTLS10,
	}
	if config.CA != "" {
		certPool := x509.NewCertPool()
		file, err := ioutil.ReadFile(config.CA)
		if err != nil {
			return nil, fmt.Errorf("Could not read CA certificate: %v", err)
		}
		certPool.AppendCertsFromPEM(file)
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		tlsConfig.ClientCAs = certPool
	}
	return tls.NewListener(l, tlsConfig), nil
}

func setSocketGroup(path, group string) error {
	if group == "" {
		return nil
	}
	if err := changeGroup(path, group); err != nil {
		if group != "docker" {
			return err
		}
		logrus.Debugf("Warning: could not change group %s to docker: %v", path, err)
	}
	return nil
}

func changeGroup(path string, nameOrGid string) error {
	gid, err := lookupGidByName(nameOrGid)
	if err != nil {
		return err
	}
	logrus.Debugf("%s group found. gid: %d", nameOrGid, gid)
	return os.Chown(path, 0, gid)
}

func lookupGidByName(nameOrGid string) (int, error) {
	groupFile, err := user.GetGroupPath()
	if err != nil {
		return -1, err
	}
	groups, err := user.ParseGroupFileFilter(groupFile, func(g user.Group) bool {
		return g.Name == nameOrGid || strconv.Itoa(g.Gid) == nameOrGid
	})
	if err != nil {
		return -1, err
	}
	if groups != nil && len(groups) > 0 {
		return groups[0].Gid, nil
	}
	gid, err := strconv.Atoi(nameOrGid)
	if err == nil {
		logrus.Warnf("Could not find GID %d", gid)
		return gid, nil
	}
	return -1, fmt.Errorf("Group %s not found", nameOrGid)
}
