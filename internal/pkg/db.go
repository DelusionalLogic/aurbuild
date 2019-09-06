package pkg

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

	data      map[int]Package
	nextIndex int

	nameIndex map[string]int
}

func NewRepo() Repo {
	return Repo{
		nextIndex: 1,
		data:      make(map[int]Package, 0),
		nameIndex: make(map[string]int, 0),
	}
}

func (r *Repo) Add(pkg *Package) (Package, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	pkg.Id = r.nextIndex
	r.nextIndex++

	r.data[pkg.Id] = *pkg
	r.nameIndex[pkg.Name] = pkg.Id

	return r.data[pkg.Id], nil
}

func (r *Repo) Update(pkg *Package) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	if pkg.Id == 0 {
		return &RepoError{message: "Package has not been added to repo"}
	}

	old := r.data[pkg.Id]
	delete(r.nameIndex, old.Name)

	r.data[pkg.Id] = *pkg
	r.nameIndex[pkg.Name] = pkg.Id

	return nil
}

func (r *Repo) ByName(name string) (*Package, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	key, prs := r.nameIndex[name]
	if !prs {
		return nil, nil
	}

	val := r.data[key]

	return &val, nil
}

func (r *Repo) List() ([]Package, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	lst := make([]Package, 0, len(r.data))
	for _, v := range r.data {
		lst = append(lst, v)
	}

	return lst, nil
}
