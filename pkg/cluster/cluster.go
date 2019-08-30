package cluster

type Cluster interface {
	Create() error
	Exists() (bool, error)
	KubeConfigPath() (string, error)
	LoadImage(string) error
	Destroy() error
}
