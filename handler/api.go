package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"syscall"

	"aurbuild/internal/pkg"
	"aurbuild/internal/pkgbuild"

	"github.com/go-chi/chi"
)

// Raise a signal on us
func raise(sig os.Signal) error {
	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		return err
	}
	return p.Signal(sig)
}

// Make the http server abort on failed requests
func PanicMiddlware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Fprintf(os.Stderr, "panic: %v", err)
				raise(syscall.SIGABRT)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func handleTime(pkgs *pkg.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		packageName := chi.URLParam(r, "name")
		pkg, err := pkgs.ByName(packageName)
		if err != nil {
			panic(err)
		}

		if pkg != nil {
			w.WriteHeader(404)
			w.Header().Set("Content-Type", "text/plain")
			return
		}

		w.WriteHeader(200)
		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, pkg.Name)
	}
}

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func ListPackages(pkgs *pkg.Repo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pkgs, err := pkgs.List()
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)

		respList := make([]Package, len(pkgs))
		for i, pkg := range pkgs {
			respList[i] = Package{
				Name:    pkg.Name,
				Version: pkg.Version,
			}
		}
		json.NewEncoder(w).Encode(respList)
	}
}

type NewPkgbuild struct {
	Body string `json:"body"`
}

type Pkgbuild struct {
	Id     int    `json:"id"`
	Body   string `json:"body"`
	Status string `json:"status"`
}

func statusRepr(s pkgbuild.Status) string {
	switch s {
	case pkgbuild.Status_Parsing:
		return "Parsing"
	case pkgbuild.Status_Parsed:
		return "Parsed"
	}

	panic("Bad status")
}

func SubmitPkgbuild(repo *pkgbuild.Repo, created chan<- pkgbuild.Created) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := NewPkgbuild{}
		json.NewDecoder(r.Body).Decode(&body)

		p, err := repo.Add(&pkgbuild.Pkgbuild{
			Body:   body.Body,
			Status: pkgbuild.Status_Parsing,
		})
		if err != nil {
			panic(err)
		}

		created <- pkgbuild.Created{
			Id: p.Id,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(Pkgbuild{
			Id:     p.Id,
			Body:   p.Body,
			Status: statusRepr(p.Status),
		})
	}
}

func Handler(pkgs *pkg.Repo, pbs *pkgbuild.Repo, created chan<- pkgbuild.Created) http.Handler {
	r := chi.NewRouter()

	r.Use(PanicMiddlware)
	r.Route("/packages", func(r chi.Router) {
		r.Get("/", ListPackages(pkgs))
	})
	r.Route("/pkgbuild", func(r chi.Router) {
		r.Post("/", SubmitPkgbuild(pbs, created))
	})

	return r
}
