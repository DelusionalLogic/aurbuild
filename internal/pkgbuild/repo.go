package pkgbuild

import (
	"sync"
)

type RepoError struct {
	message string
}

func (r *RepoError) Error() string {
	return r.message
}

type Repo struct {
	lock sync.Mutex

	data      map[int]Pkgbuild
	nextIndex int
}

func NewRepo() Repo {
	return Repo{
		nextIndex: 1,
		data:      make(map[int]Pkgbuild, 0),
	}
}

func (r *Repo) Add(pkg *Pkgbuild) (Pkgbuild, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	pkg.Id = r.nextIndex
	r.nextIndex++

	r.data[pkg.Id] = *pkg

	return r.data[pkg.Id], nil
}

func (r *Repo) Update(pkg *Pkgbuild) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.data[pkg.Id] = *pkg

	return nil
}

func (r *Repo) Get(id int) (Pkgbuild, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.data[id], nil
}
