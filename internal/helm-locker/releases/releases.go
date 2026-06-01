package releases

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/sirupsen/logrus"
	releasev1 "helm.sh/helm/v4/pkg/release/v1"
	"helm.sh/helm/v4/pkg/storage"
	"helm.sh/helm/v4/pkg/storage/driver"
	"k8s.io/client-go/kubernetes"
)

// logrusHandler bridges Helm v4's slog-based storage logger to logrus.
type logrusHandler struct {
	preAttrs []slog.Attr
}

func (h *logrusHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= slog.LevelDebug
}

func (h *logrusHandler) Handle(_ context.Context, r slog.Record) error {
	fields := make(logrus.Fields)
	for _, a := range h.preAttrs {
		fields[a.Key] = a.Value.Any()
	}
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = a.Value.Any()
		return true
	})
	entry := logrus.WithFields(fields)
	switch {
	case r.Level >= slog.LevelError:
		entry.Error(r.Message)
	case r.Level >= slog.LevelWarn:
		entry.Warn(r.Message)
	case r.Level >= slog.LevelInfo:
		entry.Info(r.Message)
	default:
		entry.Debug(r.Message)
	}
	return nil
}

func (h *logrusHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &logrusHandler{preAttrs: append(h.preAttrs, attrs...)}
}
func (h *logrusHandler) WithGroup(_ string) slog.Handler { return h }

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
	store.SetLogger(&logrusHandler{})
	g.namespacedStorage[namespace] = store
	return store
}

func (g *latestReleaseGetter) Last(namespace, name string) (*releasev1.Release, error) {
	store := g.getStore(namespace)
	rel, err := store.Last(name)
	if err != nil {
		return nil, err
	}
	switch r := rel.(type) {
	case *releasev1.Release:
		return r, nil
	case releasev1.Release:
		return &r, nil
	case nil:
		return nil, fmt.Errorf("helm storage returned nil release for %s/%s", namespace, name)
	default:
		return nil, fmt.Errorf("helm storage returned unexpected type %T for release %s/%s", rel, namespace, name)
	}
}
