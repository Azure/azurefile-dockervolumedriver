# Docker volume extension api.

Go handler to create external volume extensions for Docker.

## Usage

This library is designed to be integrated in your program.

1. Implement the `dkvolume.Driver` interface.
2. Initialize a `dkvolume.Hander` with your implementation.
3. Call either `ServeTCP` or `ServeUnix` from the `dkvolume.Handler`.

### Example using TCP sockets:

```go
  d := MyVolumeDriver{}
  h := dkvolume.NewHandler(d)
  h.ServeTCP("test_volume", ":8080")
```

### Example using Unix sockets:

```go
  d := MyVolumeDriver{}
  h := dkvolume.NewHandler(d)
  h.ServeUnix("root", "test_volume")
```

## Full example plugins

- https://github.com/calavera/docker-volume-glusterfs
- https://github.com/calavera/docker-volume-keywhiz

## License

MIT
