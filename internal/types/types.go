package types

type Identifier struct {
	Name     string
	Category string
}

type ContainerResult struct {
	Identifier  Identifier
	Output      string
	BuildOutput string
	Error       error
}

type ContainerModule struct {
	Identifier Identifier
	Dockerfile string
	Command    []string
}
