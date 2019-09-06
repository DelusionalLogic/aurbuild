package pkgbuild

type Status int

const (
	Status_Parsing Status = iota
	Status_Parsed
)

type Pkgbuild struct {
	Id     int
	Body   string
	Status Status
}

type Created struct {
	Id int
}

type Parsed struct {
	Id      int
	Name    string
	Version string
}
