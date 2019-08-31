package cluster

import "fmt"

// ErrNotExists returned if cluster with requested name doesn't exist
type ErrNotExists struct {
	Name string
}

func (err ErrNotExists) Error() string {
	return fmt.Sprintf("cluster: %s not exists", err.Name)
}

// ErrAlreadyExists returned if cluster with the specific name has been already created
type ErrAlreadyExists struct {
	Name string
}

func (err ErrAlreadyExists) Error() string {
	return fmt.Sprintf("cluster: %s already exists", err.Name)
}
