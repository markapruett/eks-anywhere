package createvalidations

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/features"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// PreflightValidations returns the validations required before creating a cluster.
func (v *CreateValidations) PreflightValidations(ctx context.Context) []validations.Validation {
	k := v.Opts.Kubectl

	targetCluster := &types.Cluster{
		Name:           v.Opts.WorkloadCluster.Name,
		KubeconfigFile: v.Opts.ManagementCluster.KubeconfigFile,
	}

	createValidations := []validations.Validation{
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate certificate for registry mirror",
				Remediation: fmt.Sprintf("provide a valid certificate for you registry endpoint using %s env var", anywherev1.RegistryMirrorCAKey),
				Err:         validations.ValidateCertForRegistryMirror(v.Opts.Spec, v.Opts.TlsValidator),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate authentication for git provider",
				Remediation: fmt.Sprintf("ensure %s, %s env variable are set and valid", config.EksaGitPrivateKeyTokenEnv, config.EksaGitKnownHostsFileEnv),
				Err:         validations.ValidateAuthenticationForGitProvider(v.Opts.Spec, v.Opts.CliConfig),
			}
		},
		func() *validations.ValidationResult {
			return &validations.ValidationResult{
				Name:        "validate kubernetes version 1.25 support",
				Remediation: fmt.Sprintf("ensure %v env variable is set", features.K8s125SupportEnvVar),
				Err:         validations.ValidateK8s125Support(v.Opts.Spec),
				Silent:      true,
			}
		},
	}

	if v.Opts.Spec.Cluster.IsManaged() {
		createValidations = append(
			createValidations,
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate cluster name",
					Remediation: "",
					Err:         ValidateClusterNameIsUnique(ctx, k, targetCluster, v.Opts.Spec.Cluster.Name),
				}
			},
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate gitops",
					Remediation: "",
					Err:         ValidateGitOps(ctx, k, v.Opts.ManagementCluster, v.Opts.Spec),
				}
			},
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate identity providers' name",
					Remediation: "",
					Err:         ValidateIdentityProviderNameIsUnique(ctx, k, targetCluster, v.Opts.Spec),
				}
			},
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name:        "validate management cluster has eksa crds",
					Remediation: "",
					Err:         ValidateManagementCluster(ctx, k, targetCluster),
				}
			},
			func() *validations.ValidationResult {
				return &validations.ValidationResult{
					Name: "validate management cluster name is valid",
					Remediation: "Specify a valid management cluster in the cluster spec. This cannot be a workload cluster that is managed by a different " +
						"management cluster.",
					Err: validations.ValidateManagementClusterName(ctx, k, v.Opts.ManagementCluster, v.Opts.Spec.Cluster.Spec.ManagementCluster.Name),
				}
			},
		)
	}

	return createValidations
}
