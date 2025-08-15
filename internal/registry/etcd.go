package registry

import (
	"github.com/go-kratos/kratos-layout/internal/conf"
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/log"
	etcdclient "go.etcd.io/etcd/client/v3"
)

func NewEtcdRegistry(bc *conf.Bootstrap, logger log.Logger) *etcd.Registry {
	client, err := etcdclient.New(etcdclient.Config{Endpoints: bc.GetRegistry().GetEndpoint()})
	if err != nil {
		panic(err)
	}

	return etcd.New(client)
}
