package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	azure "github.com/Azure/azure-sdk-for-go/storage"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

type volumeDriver struct {
	m            sync.Mutex
	cl           azure.FileServiceClient
	meta         *metadataDriver
	accountName  string
	accountKey   string
	storageBase  string
	mountpoint   string
	removeShares bool
}

func newVolumeDriver(accountName, accountKey, storageBase, mountpoint, metadataRoot string, removeShares bool) (*volumeDriver, error) {
	storageClient, err := azure.NewClient(accountName, accountKey, storageBase, azure.DefaultAPIVersion, true)
	if err != nil {
		return nil, fmt.Errorf("error creating azure client: %v", err)
	}
	metaDriver, err := newMetadataDriver(metadataRoot)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize metadata driver: %v", err)
	}
	return &volumeDriver{
		cl:           storageClient.GetFileService(),
		meta:         metaDriver,
		accountName:  accountName,
		accountKey:   accountKey,
		storageBase:  storageBase,
		mountpoint:   mountpoint,
		removeShares: removeShares,
	}, nil
}

func (v *volumeDriver) Create(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "create",
		"name":      req.Name,
		"options":   req.Options})

	volMeta, err := v.meta.Validate(req.Options)
	if err != nil {
		resp.Err = fmt.Sprintf("error validating metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}

	// Additional volume metadata
	volMeta.Account = v.accountName
	volMeta.CreatedAt = time.Now().UTC()

	share := req.Options["share"]
	if share == "" {
		resp.Err = "missing volume option: 'share'"
		logctx.Error(resp.Err)
		return
	}

	logctx.Debug("request accepted")

	// Create azure file share
	if ok, err := v.cl.CreateShareIfNotExists(share); err != nil {
		resp.Err = fmt.Sprintf("error creating azure file share: %v", err)
		logctx.Error(resp.Err)
		return
	} else if ok {
		logctx.Infof("created azure file share %q", share)
	}

	// Save volume metadata
	if err := v.meta.Set(req.Name, volMeta); err != nil {
		resp.Err = fmt.Sprintf("error saving metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}
	return
}

func (v *volumeDriver) Path(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	log.WithFields(log.Fields{
		"operation": "path", "name": req.Name,
	}).Debug("request accepted")

	resp.Mountpoint = v.pathForVolume(req.Name)
	return
}

func (v *volumeDriver) Mount(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "mount",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	path := v.pathForVolume(req.Name)
	if err := os.MkdirAll(path, 0700); err != nil {
		resp.Err = fmt.Sprintf("could not create mount point: %v", err)
		logctx.Error(resp.Err)
		return
	}

	meta, err := v.meta.Get(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if meta.Account != v.accountName {
		resp.Err = fmt.Sprintf("volume hosted on a different account ('%s') cannot mount", meta.Account)
		logctx.Error(resp.Err)
		return
	}

	if err := mount(v.accountName, v.accountKey, v.storageBase, path, meta.Options); err != nil {
		resp.Err = err.Error()
		logctx.Error(resp.Err)
		return
	}
	resp.Mountpoint = path
	return
}

func (v *volumeDriver) Unmount(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "unmount",
		"name":      req.Name,
	})

	logctx.Debug("request accepted")
	path := v.pathForVolume(req.Name)
	if err := unmount(path); err != nil {
		resp.Err = err.Error()
		logctx.Error(resp.Err)
		return
	}
	logctx.Debug("unmount successful")

	// Docker does not keep track of what is mounted and what is not, it will
	// issue /Volume.Mount and /Volume.Unmount requests regardless when multiple
	// containers use the same volume simulatenosly. This leads to duplicate
	// mount entries and requirement for a careful cleanup of the mountpath in
	// the following code.
	//
	// If same path is mounted multiple times, duplicate entries will occur
	// in mount table for the same mountpoint. umount will remove the mount
	// entry but the mountpoint will still be active (and mounted).
	//
	// In that case, we read the mount table to see if there is still something
	// mounted, and only when there is nothing mounted, we remove the mountpoint
	isActive, err := isMounted(path)
	if err != nil {
		resp.Err = err.Error()
		logctx.Error(resp.Err)
		return
	}
	if isActive {
		logctx.Debug("mountpoint still has active mounts, not removing")
	} else {
		logctx.Debug("mountpoint has no further mounts, removing")
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			resp.Err = fmt.Sprintf("error removing mountpoint: %v", err)
			logctx.Error(resp.Err)
			return
		}
	}
	return
}

func (v *volumeDriver) Remove(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "remove",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	meta, err := v.meta.Get(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}

	share := meta.Options.Share
	if v.removeShares {
		if ok, err := v.cl.DeleteShareIfExists(share); err != nil {
			resp.Err = fmt.Sprintf("error removing azure file share %q: %v", share, err)
			logctx.Error(resp.Err)
			return
		} else if ok {
			logctx.Infof("removed azure file share %q", share)
		}
	} else {
		logctx.Debugf("not removing share %q upon volume removal", share)
	}

	logctx.Debug("removing volume metadata")
	if err != v.meta.Delete(req.Name) {
		resp.Err = err.Error()
		logctx.Error(resp.Err)
		return
	}
	return
}

func (v *volumeDriver) Get(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()
	logctx := log.WithFields(log.Fields{
		"operation": "get",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	_, err := v.meta.Get(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}
	resp.Volume = v.volumeEntry(req.Name)
	return
}

func (v *volumeDriver) List(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "list",
	})
	logctx.Debug("request accepted")

	vols, err := v.meta.List()
	if err != nil {
		resp.Err = fmt.Sprintf("failed to list managed volumes: %v", err)
		logctx.Error(resp.Err)
		return
	}

	for _, vn := range vols {
		resp.Volumes = append(resp.Volumes, v.volumeEntry(vn))
	}
	logctx.Debugf("response has %d items", len(resp.Volumes))
	return
}

func (v *volumeDriver) volumeEntry(name string) *volume.Volume {
	return &volume.Volume{Name: name,
		Mountpoint: v.pathForVolume(name)}
}

func (v *volumeDriver) pathForVolume(name string) string {
	return filepath.Join(v.mountpoint, name)
}

func mount(accountName, accountKey, storageBase, mountPath string, options VolumeOptions) error {
	// Set defaults
	if len(options.FileMode) == 0 {
		options.FileMode = "0777"
	}
	if len(options.DirMode) == 0 {
		options.DirMode = "0777"
	}
	if len(options.UID) == 0 {
		options.UID = "0"
	}
	if len(options.GID) == 0 {
		options.GID = "0"
	}
	mountURI := fmt.Sprintf("//%s.file.%s/%s", accountName, storageBase, options.Share)
	opts := []string{
		"vers=3.0",
		fmt.Sprintf("username=%s", accountName),
		fmt.Sprintf("password=%s", accountKey),
		fmt.Sprintf("file_mode=%s", options.FileMode),
		fmt.Sprintf("dir_mode=%s", options.DirMode),
		fmt.Sprintf("uid=%s", options.UID),
		fmt.Sprintf("gid=%s", options.GID),
	}
	if options.NoLock {
		opts = append(opts, "nolock")
	}

	// TODO: replace with mount() syscall using docker/docker/pkg/mount
	// (currently gives hard-to-debug 'invalid argument' error with the
	// following arguments, my guess is, mount program does IP resolution
	// and essentially passes a different set of options to system call).
	cmd := exec.Command("mount", "-t", "cifs", mountURI, mountPath, "-o", strings.Join(opts, ","), "--verbose")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %v\noutput=%q", err, out)
	}
	return nil
}

func unmount(mountpoint string) error {
	cmd := exec.Command("umount", mountpoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount failed: %v\noutput=%q", err, out)
	}
	return nil
}

// isMounted reads /proc/self/mountinfo to see if the specified mountpoint is
// mounted.
func isMounted(mountpoint string) (bool, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return false, fmt.Errorf("cannot read mountinfo: %v", err)
	}
	defer f.Close()

	// format of mountinfo:
	//    38 23 0:30 / /sys/fs/cgroup/devices rw,relatime - cgroup cgroup rw,devices
	//    39 23 0:31 / /sys/fs/cgroup/freezer rw,relatime - cgroup cgroup rw,freezer
	//    33 22 8:17 / /mnt rw,relatime - ext4 /dev/sdb1 rw,data=ordered
	// so we split the lines into the specified format and match the mountpoint
	// at 5th field.
	//
	// This code is adopted from https://github.com/docker/docker/blob/master/pkg/mount/mountinfo_linux.go

	oldFi, err := os.Stat(mountpoint)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("cannot stat mountpoint: %v", err)
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		t := s.Text()
		f := strings.Fields(t)
		if len(f) < 5 {
			return false, fmt.Errorf("mountinfo line %q has less than 5 fields, cannot parse mountpoint", t)
		}
		mp := f[4] // ID, Parent, Major, Minor, Root, *Mountpoint*, Opts, OptionalFields
		fi, err := os.Stat(mp)
		if err != nil {
			return false, fmt.Errorf("cannot stat %s: %v", mp, err)
		}
		same := os.SameFile(oldFi, fi)
		if same {
			return true, nil
		}
	}
	log.Debug("mountpoint not found")
	return false, nil
}
