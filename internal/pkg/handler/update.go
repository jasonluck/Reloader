package handler

import (
	"github.com/sirupsen/logrus"
	"github.com/stakater/Reloader/internal/pkg/metrics"
	"github.com/stakater/Reloader/internal/pkg/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	secretsstorev1 "sigs.k8s.io/secrets-store-csi-driver/apis/v1"
)

// ResourceUpdatedHandler contains updated objects
type ResourceUpdatedHandler struct {
	Resource    interface{}
	OldResource interface{}
	Collectors  metrics.Collectors
	Recorder    record.EventRecorder
}

// Handle processes the updated resource
func (r ResourceUpdatedHandler) Handle() error {
	if r.Resource == nil || r.OldResource == nil {
		logrus.Errorf("Resource update handler received nil resource")
	} else {
		config, oldSHAData := r.GetConfig()
		if config.SHAValue != oldSHAData {
			// process resource based on its type
			return doRollingUpgrade(config, r.Collectors, r.Recorder)
		}
	}
	return nil
}

// GetConfig gets configurations containing SHA, annotations, namespace and resource name
func (r ResourceUpdatedHandler) GetConfig() (util.Config, string) {
	var oldSHAData string
	var config util.Config
	if _, ok := r.Resource.(*v1.ConfigMap); ok {
		oldSHAData = util.GetSHAfromConfigmap(r.OldResource.(*v1.ConfigMap))
		config = util.GetConfigmapConfig(r.Resource.(*v1.ConfigMap))
	} else if _, ok := r.Resource.(*v1.Secret); ok {
		oldSHAData = util.GetSHAfromSecret(r.OldResource.(*v1.Secret).Data)
		config = util.GetSecretConfig(r.Resource.(*v1.Secret))
	} else if _, ok := r.Resource.(*secretsstorev1.SecretProviderClass); ok {
		oldSHAData = util.GetSHAfromSecret(r.OldResource.(*secretsstorev1.SecretProviderClass))
		config = util.GetSecretConfig(r.Resource.(*secretsstorev1.SecretProviderClass))
	} else {
		logrus.Warnf("Invalid resource: Resource should be 'Secret' or 'Configmap' but found, %v", r.Resource)
	}
	return config, oldSHAData
}
