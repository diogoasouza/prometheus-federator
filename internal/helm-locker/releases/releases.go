package releases

import (
	"fmt"
	"sync"

	releasev1 "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
)

type HelmReleaseGetter interface {
	Last(namespace, name string) (*releasev1.Release, error)
}

func NewHelmReleaseGetter(k8s kubernetes.Interface) HelmReleaseGetter {
	return &latestReleaseGetter{
		K8s:               k8s,
		namespacedStorage: make(map[string]*storage.Storage),
	}
}

type latestReleaseGetter struct {
	K8s kubernetes.Interface

	namespacedStorage map[string]*storage.Storage
	storageLock       sync.Mutex
}

func (g *latestReleaseGetter) getStore(namespace string) *storage.Storage {
	g.storageLock.Lock()
	defer g.storageLock.Unlock()
	store, ok := g.namespacedStorage[namespace]
	if ok && store != nil {
		return store
	}
	store = storage.Init(driver.NewSecrets(g.K8s.CoreV1().Secrets(namespace)))
	g.namespacedStorage[namespace] = store
	return store
}

func (g *latestReleaseGetter) Last(namespace, name string) (*releasev1.Release, error) {
	store := g.getStore(namespace)
	rel, err := store.Last(name)
	if err != nil {
		return nil, err
	}
	r, ok := rel.(*releasev1.Release)
	if !ok {
		return nil, fmt.Errorf("unexpected release type %T", rel)
	}
	return r, nil
}
