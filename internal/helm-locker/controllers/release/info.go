package release

import (
	v1alpha1 "github.com/rancher/prometheus-federator/internal/helm-locker/apis/helm.cattle.io/v1alpha1"
	"helm.sh/helm/v4/pkg/release/common"
	releasev1 "helm.sh/helm/v4/pkg/release/v1"
)

func newReleaseInfo(release *releasev1.Release) *releaseInfo {
	info := &releaseInfo{}
	info.Version = int(release.Version)
	info.Manifest = release.Manifest
	if release.Info != nil {
		info.Description = release.Info.Description
		info.Notes = release.Info.Notes
		switch release.Info.Status {
		case common.StatusUnknown:
			info.State = v1alpha1.UnknownState
		case common.StatusDeployed:
			info.State = v1alpha1.DeployedState
		case common.StatusUninstalled:
			info.State = v1alpha1.UninstalledState
		case common.StatusSuperseded:
			// note: this should never be the case since we always get the latest secret
			info.State = v1alpha1.ErrorState
		case common.StatusFailed:
			info.State = v1alpha1.FailedState
		default:
			// uninstalling, pending install, pending upgrade, pending rollback
			info.State = v1alpha1.TransitioningState
		}
	}
	return info
}

type releaseInfo struct {
	Version     int
	Manifest    string
	Description string
	Notes       string
	State       string
}

func (i *releaseInfo) Locked() bool {
	return i.State == v1alpha1.DeployedState
}

func (i *releaseInfo) GetUpdatedStatus(helmRelease *v1alpha1.HelmRelease) *v1alpha1.HelmRelease {
	helmRelease.Status.Version = i.Version
	helmRelease.Status.Description = i.Description
	helmRelease.Status.State = i.State
	helmRelease.Status.Notes = i.Notes
	return helmRelease
}
