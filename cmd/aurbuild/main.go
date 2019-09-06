package main

import (
	"net/http"
	"os"

	"aurbuild/handler"
	"aurbuild/internal/build"
	"aurbuild/internal/bus"
	"aurbuild/internal/pkg"
	"aurbuild/internal/pkgbuild"

	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"golang.org/x/sys/unix"
)

func main() {
	ctx := context.Background()

	// @HACK
	unix.Umask(0000)

	err := os.Mkdir("/tmp/aurbuild", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	err = os.Mkdir("/tmp/aurbuild/workspaces", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		panic(err)
	}

	pkgs := pkg.NewRepo()
	pbs := pkgbuild.NewRepo()

	created := make(chan pkgbuild.Created)
	parsed := bus.MakeQueue()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	parsed.Subscribe(func(payload interface{}) {
		event := payload.(pkgbuild.Parsed)

		pb, err := pbs.Get(event.Id)
		if err != nil {
			panic(err)
		}

		pb.Status = pkgbuild.Status_Parsed

		err = pbs.Update(&pb)
		if err != nil {
			panic(err)
		}
	})

	parsed.Subscribe(func(payload interface{}) {
		event := payload.(pkgbuild.Parsed)

		pkgs.Add(&pkg.Package{
			Name:    event.Name,
			Version: event.Version,
		})
	})

	go func() {
		for {
			event, ok := <-created
			if !ok {
				panic("Created loop broke")
			}
			print("Created\n")

			pb, err := pbs.Get(event.Id)
			if err != nil {
				panic(err)
			}

			ws, err := build.FromString("/tmp/aurbuild/workspaces", pb.Body)
			if err != nil {
				panic(err)
			}

			info, err := build.Parse(cli, ctx, ws)
			if err != nil {
				panic(err)
			}

			parsed.Post(pkgbuild.Parsed{
				Id:      event.Id,
				Name:    info.Name,
				Version: info.Version,
			})
		}
	}()

	r := api.Handler(&pkgs, &pbs, created)

	panic(http.ListenAndServe(":8080", r))

	/*
		ws, err := pkg.Load("/tmp/aurbuild/workspaces", "./testroot/PKGBUILD")
		if err != nil {
			panic(err)
		}
		defer ws.Close()

		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}

		info, err := pkg.Parse(cli, ctx, ws)
		if err != nil {
			panic(err)
		}

		println("Building: ", info.Name, info.Version)

		err = pkg.Build(cli, ctx, ws, &info.Depends)
		if err != nil {
			panic(err)
		}
	*/
}
