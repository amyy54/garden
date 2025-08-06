package modules

type Module struct {
	Name    string   `json:"name" validate:"required"`
	Command []string `json:"cmd" validate:"required"`
	Hash    string   `json:"sha256" validate:"required"`
}

type Category struct {
	Version int      `json:"version" validate:"required"`
	Name    string   `validate:"required"`
	Modules []Module `json:"modules" validate:"required"`
}

type CategoryConfig struct {
	Name string `json:"name" validate:"required"`
	Hash string `json:"sha256" validate:"required"`
}

type Meta struct {
	Version    int              `json:"version" validate:"required"`
	Categories []CategoryConfig `json:"categories" validate:"required"`
	Path       string           `validate:"required"`
}

type ModuleOptions struct {
	Categories   []string
	Modules      []string
	IgnoreHashes bool
}
