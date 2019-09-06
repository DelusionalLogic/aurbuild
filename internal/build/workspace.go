package build

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

type Workspace struct {
	Root     string
	PkgRoot  string
	ParseDir string
	BuildDir string
}

func (w *Workspace) Close() error {
	err := os.Remove(w.Root)
	if err != nil {
		return err
	}

	os.Remove(w.PkgRoot)
	if err != nil {
		return err
	}

	os.Remove(w.ParseDir)
	if err != nil {
		return err
	}

	os.Remove(w.BuildDir)
	if err != nil {
		return err
	}

	return nil
}

func createWorkspace(workdir string) (*string, *string, *string, error) {
	// Create workspace
	workdir, err := ioutil.TempDir(workdir, "")
	if err != nil {
		return nil, nil, nil, err
	}

	// Create package root folder
	rootdir := path.Join(workdir, "pkgroot")
	err = os.Mkdir(rootdir, 0777)
	if err != nil {

		err = os.Remove(workdir)
		if err != nil {
			panic(err)
		}

		return nil, nil, nil, err
	}

	parsedir := path.Join(workdir, "parse")
	err = os.Mkdir(parsedir, 0777)
	if err != nil {
		err = os.Remove(rootdir)
		if err != nil {
			panic(err)
		}

		err = os.Remove(workdir)
		if err != nil {
			panic(err)
		}
		return nil, nil, nil, err
	}

	builddir := path.Join(workdir, "build")
	err = os.Mkdir(builddir, 0777)
	if err != nil {

		err = os.Remove(parsedir)
		if err != nil {
			panic(err)
		}

		err = os.Remove(rootdir)
		if err != nil {
			panic(err)
		}

		err = os.Remove(workdir)
		if err != nil {
			panic(err)
		}
		return nil, nil, nil, err
	}

	return &rootdir, &parsedir, &builddir, nil
}

func FromString(workdir, body string) (*Workspace, error) {
	rootdir, parsedir, builddir, err := createWorkspace(workdir)
	if err != nil {
		return nil, err
	}

	to, err := os.OpenFile(path.Join(*rootdir, "PKGBUILD"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer to.Close()

	_, err = to.WriteString(body)
	if err != nil {
		return nil, err
	}

	return &Workspace{
		Root:     workdir,
		PkgRoot:  *rootdir,
		ParseDir: *parsedir,
		BuildDir: *builddir,
	}, nil
}

func FromFile(workdir, pkgbuild string) (*Workspace, error) {
	rootdir, parsedir, builddir, err := createWorkspace(workdir)
	if err != nil {
		return nil, err
	}

	// Copy file
	from, err := os.Open(pkgbuild)
	if err != nil {
		return nil, err
	}
	defer from.Close()

	to, err := os.OpenFile(path.Join(*rootdir, "PKGBUILD"), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return nil, err
	}

	return &Workspace{
		Root:     workdir,
		PkgRoot:  *rootdir,
		ParseDir: *parsedir,
		BuildDir: *builddir,
	}, nil
}
