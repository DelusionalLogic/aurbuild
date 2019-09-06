package build

import (
	"encoding/json"
	"os"
	"path"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"golang.org/x/net/context"
)

type Deps struct {
	Install []string `json:"install"`
	Build   []string `json:"build"`
	Check   []string `json:"check"`
}

type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Depends Deps   `json:"depends"`
}

func Parse(cli client.APIClient, ctx context.Context, ws *Workspace) (*Info, error) {
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "archparse",
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			mount.Mount{
				Type:   mount.TypeBind,
				Source: ws.PkgRoot,
				Target: "/pkgroot/",
			},
			mount.Mount{
				Type:   mount.TypeBind,
				Source: ws.ParseDir,
				Target: "/out/",
			},
		},
	}, nil, "")
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		return nil, err
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	finfo, err := os.Open(path.Join(ws.ParseDir, "info.json"))
	if err != nil {
		return nil, err
	}
	defer finfo.Close()

	var info Info
	dec := json.NewDecoder(finfo)
	err = dec.Decode(&info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func Build(cli client.APIClient, ctx context.Context, ws *Workspace, dep *Deps) error {
	depSize := (len(dep.Install) + len(dep.Build) + len(dep.Check))
	cmd := make([]string, depSize*2)

	offset := 0
	for i := 0; i < len(dep.Install); i++ {
		cmd[offset+i*2] = "-i"
		cmd[offset+i*2+1] = dep.Install[i]
	}
	offset = len(dep.Install) * 2

	for i := 0; i < len(dep.Build); i++ {
		cmd[offset+i*2] = "-i"
		cmd[offset+i*2+1] = dep.Build[i]
	}
	offset = len(dep.Build) * 2

	for i := 0; i < len(dep.Check); i++ {
		cmd[offset+i*2] = "-i"
		cmd[offset+i*2+1] = dep.Check[i]
	}
	offset = len(dep.Check) * 2

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "archbuild",
		Cmd:   cmd,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			mount.Mount{
				Type:   mount.TypeBind,
				Source: ws.PkgRoot,
				Target: "/pkgroot/",
			},
			mount.Mount{
				Type:   mount.TypeBind,
				Source: ws.BuildDir,
				Target: "/out/",
			},
		},
	}, nil, "")
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	attach, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdout: true,
		Stderr: true,
	})

	stdcopy.StdCopy(os.Stdout, os.Stderr, attach.Reader)

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-statusCh:
	}

	attach.Close()

	return nil
}
