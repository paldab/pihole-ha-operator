package conditions

const (
	ReasonClusterReady = "ClusterReady"

	ReasonReconciling          = "Reconciling"
	ReasonFailingOver          = "FailingOver"
	ReasonStatefulSetNotReady  = "StatefulSetNotReady"
	ReasonStatefulSetNotFound  = "StatefulSetNotFound"
	ReasonImmutableFieldChange = "ImmutableStatefulSetChange"

	ReasonConfigurationValid   = "ConfigurationValid"
	ReasonConfigurationInvalid = "ConfigurationInvalid"

	ReasonSecretNotFound = "SecretNotFound"
	ReasonServiceFailed  = "ServiceReconcileFailed"
	ReasonIngressFailed  = "IngressReconcileFailed"
)
