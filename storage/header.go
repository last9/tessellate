package storage

import (
	"github.com/hashicorp/consul/api"
	"github.com/tsocial/tessellate/storage/types"
)

type Storer interface {
	Setup() error
	Teardown() error
	GetClient() *api.Client

	Save(reader types.ReaderWriter, tree *types.Tree) error
	Get(reader types.ReaderWriter, tree *types.Tree) error
	GetVersion(reader types.ReaderWriter, tree *types.Tree, version string) error
	GetVersions(reader types.ReaderWriter, tree *types.Tree) ([]string, error)
}
