package dkvolume

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// DefaultDockerRootDirectory is the default directory where volumes will be created.
	DefaultDockerRootDirectory = "/var/lib/docker-volumes"

	defaultContentTypeV1_1        = "application/vnd.docker.plugins.v1.1+json"
	defaultImplementationManifest = `{"Implements": ["VolumeDriver"]}`
	pluginSpecDir                 = "/etc/docker/plugins"
	pluginSockDir                 = "/run/docker/plugins"

	activatePath    = "/Plugin.Activate"
	createPath      = "/VolumeDriver.Create"
	removePath      = "/VolumeDriver.Remove"
	hostVirtualPath = "/VolumeDriver.Path"
	mountPath       = "/VolumeDriver.Mount"
	unmountPath     = "/VolumeDriver.Unmount"
)

// Driver represent the interface a driver must fulfill.
type Driver interface {
	Create(CreateRequest) (CreateResponse, error)
	Remove(RemoveRequest) (RemoveResponse, error)
	Path(PathRequest) (PathResponse, error)
	Mount(MountRequest) (MountResponse, error)
	Unmount(UnmountRequest) (UnmountResponse, error)
}

type errorResponse struct {
	Err string `json:"Err"`
}

type CreateRequest struct {
	Name    string            `json:"Name"`
	Options map[string]string `json:"Opts"`
}

type CreateResponse struct{}

type RemoveRequest struct {
	Name string `json:"Name"`
}

type RemoveResponse struct{}

type PathRequest struct {
	Name string `json:"Name"`
}

type PathResponse struct {
	Mountpoint string `json:"Mountpoint"`
}

type MountRequest struct {
	Name string `json:"Name"`
}

type MountResponse struct {
	Mountpoint string `json:"Mountpoint"`
}

type UnmountRequest struct {
	Name string `json:"Name"`
}

type UnmountResponse struct{}

// Handler forwards requests and responses between the docker daemon and the plugin.
type Handler struct {
	driver Driver
	mux    *http.ServeMux
}

// NewHandler initializes the request handler with a driver implementation.
func NewHandler(driver Driver) *Handler {
	h := &Handler{driver, http.NewServeMux()}
	h.initMux()
	return h
}

func (h *Handler) initMux() {
	h.mux.HandleFunc(activatePath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", defaultContentTypeV1_1)
		fmt.Fprintln(w, defaultImplementationManifest)
	})

	h.mux.HandleFunc(createPath, func(w http.ResponseWriter, r *http.Request) {
		var req CreateRequest
		if err := decodeRequest(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.driver.Create(req)
		writeResponse(w, res, err)
	})

	h.mux.HandleFunc(removePath, func(w http.ResponseWriter, r *http.Request) {
		var req RemoveRequest
		if err := decodeRequest(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.driver.Remove(req)
		writeResponse(w, res, err)
	})

	h.mux.HandleFunc(hostVirtualPath, func(w http.ResponseWriter, r *http.Request) {
		var req PathRequest
		if err := decodeRequest(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.driver.Path(req)
		writeResponse(w, res, err)
	})

	h.mux.HandleFunc(mountPath, func(w http.ResponseWriter, r *http.Request) {
		var req MountRequest
		if err := decodeRequest(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.driver.Mount(req)
		writeResponse(w, res, err)
	})

	h.mux.HandleFunc(unmountPath, func(w http.ResponseWriter, r *http.Request) {
		var req UnmountRequest
		if err := decodeRequest(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		res, err := h.driver.Unmount(req)
		writeResponse(w, res, err)
	})
}

// ServeTCP makes the handler to listen for request in a given TCP address.
// It also writes the spec file on the right directory for docker to read.
func (h *Handler) ServeTCP(pluginName, addr string) error {
	return h.listenAndServe("tcp", addr, pluginName)
}

// ServeUnix makes the handler to listen for requests in a unix socket.
// It also creates the socket file on the right directory for docker to read.
func (h *Handler) ServeUnix(systemGroup, addr string) error {
	return h.listenAndServe("unix", addr, systemGroup)
}

func (h *Handler) listenAndServe(proto, addr, group string) error {
	var (
		start = make(chan struct{})
		l     net.Listener
		err   error
		spec  string
	)

	server := http.Server{
		Addr:    addr,
		Handler: h.mux,
	}

	switch proto {
	case "tcp":
		l, err = newTCPSocket(addr, nil, start)
		if err == nil {
			spec, err = writeSpec(group, l.Addr().String())
		}
	case "unix":
		spec, err = fullSocketAddr(addr)
		if err == nil {
			l, err = newUnixSocket(spec, group, start)
		}
	}

	if spec != "" {
		defer os.Remove(spec)
	}
	if err != nil {
		return err
	}

	close(start)
	return server.Serve(l)
}

func decodeRequest(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeResponse(w http.ResponseWriter, res interface{}, err error) {
	w.Header().Set("Content-Type", defaultContentTypeV1_1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(errorResponse{err.Error()})
	} else {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func writeSpec(name, addr string) (string, error) {
	if err := os.MkdirAll(pluginSpecDir, 0755); err != nil {
		return "", err
	}

	spec := filepath.Join(pluginSpecDir, name+".spec")
	url := "tcp://" + addr
	if err := ioutil.WriteFile(spec, []byte(url), 0644); err != nil {
		return "", err
	}
	return spec, nil
}

func fullSocketAddr(addr string) (string, error) {
	if err := os.MkdirAll(pluginSockDir, 0755); err != nil {
		return "", err
	}

	if filepath.IsAbs(addr) {
		return addr, nil
	}

	return filepath.Join(pluginSockDir, addr+".sock"), nil
}
