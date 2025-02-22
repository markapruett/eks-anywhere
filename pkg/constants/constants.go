package constants

// Namespace constants.
const (
	EksaSystemNamespace                     = "eksa-system"
	EksaDiagnosticsNamespace                = "eksa-diagnostics"
	EksaControllerManagerDeployment         = "eksa-controller-manager"
	CapdSystemNamespace                     = "capd-system"
	CapcSystemNamespace                     = "capc-system"
	CapiKubeadmBootstrapSystemNamespace     = "capi-kubeadm-bootstrap-system"
	CapiKubeadmControlPlaneSystemNamespace  = "capi-kubeadm-control-plane-system"
	CapiSystemNamespace                     = "capi-system"
	CapiWebhookSystemNamespace              = "capi-webhook-system"
	CapvSystemNamespace                     = "capv-system"
	CaptSystemNamespace                     = "capt-system"
	CapaSystemNamespace                     = "capa-system"
	CapasSystemNamespace                    = "capas-system"
	CapxSystemNamespace                     = "capx-system"
	CertManagerNamespace                    = "cert-manager"
	DefaultNamespace                        = "default"
	EtcdAdmBootstrapProviderSystemNamespace = "etcdadm-bootstrap-provider-system"
	EtcdAdmControllerSystemNamespace        = "etcdadm-controller-system"
	KubeNodeLeaseNamespace                  = "kube-node-lease"
	KubePublicNamespace                     = "kube-public"
	KubeSystemNamespace                     = "kube-system"
	LocalPathStorageNamespace               = "local-path-storage"
	EtcdAdmBootstrapProviderName            = "bootstrap-etcdadm-bootstrap"
	EtcdadmControllerProviderName           = "bootstrap-etcdadm-controller"
	DefaultHttpsPort                        = "443"
	DefaultWorkerNodeGroupName              = "md-0"
	DefaultNodeCidrMaskSize                 = 24

	VSphereProviderName    = "vsphere"
	DockerProviderName     = "docker"
	AWSProviderName        = "aws"
	SnowProviderName       = "snow"
	TinkerbellProviderName = "tinkerbell"
	CloudStackProviderName = "cloudstack"
	NutanixProviderName    = "nutanix"

	VSphereCredentialsName = "vsphere-credentials"
	EksaLicenseName        = "eksa-license"
	EksaPackagesName       = "eksa-packages"

	CloudstackAnnotationSuffix = "cloudstack.anywhere.eks.amazonaws.com/v1alpha1"

	FailureDomainLabelName = "cluster.x-k8s.io/failure-domain"

	// CloudstackFailureDomainPlaceholder Provider specific keywork placeholder.
	CloudstackFailureDomainPlaceholder = "ds.meta_data.failuredomain"

	// DefaultCoreEKSARegistry is the default registry for eks-a core artifacts.
	DefaultCoreEKSARegistry = "public.ecr.aws"
	// DefaultCuratedPackagesRegistryRegex matches the default registry for curated packages in all regions.
	DefaultCuratedPackagesRegistryRegex = "783794618700.dkr.ecr.*.amazonaws.com"

	// Provider specific env vars.
	VSphereUsernameKey     = "VSPHERE_USERNAME"
	VSpherePasswordKey     = "VSPHERE_PASSWORD"
	GovcUsernameKey        = "GOVC_USERNAME"
	GovcPasswordKey        = "GOVC_PASSWORD"
	SnowCredentialsKey     = "AWS_B64ENCODED_CREDENTIALS"
	SnowCertsKey           = "AWS_B64ENCODED_CA_BUNDLES"
	NutanixUsernameKey     = "NUTANIX_USER"
	NutanixPasswordKey     = "NUTANIX_PASSWORD"
	EksaNutanixUsernameKey = "EKSA_NUTANIX_USERNAME"
	EksaNutanixPasswordKey = "EKSA_NUTANIX_PASSWORD"

	SecretKind             = "Secret"
	ConfigMapKind          = "ConfigMap"
	ClusterResourceSetKind = "ClusterResourceSet"
)

type Operation int

const (
	Create  Operation = 0
	Upgrade Operation = 1
	Delete  Operation = 2
)
