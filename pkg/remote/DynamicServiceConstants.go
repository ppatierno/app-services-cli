package remote

// DynamicServiceConstants is a struct that contains the remote service constants
type DynamicServiceConstants struct {
	Version int `json:"version"`
	Kafka   struct {
		Ams AmsConfig `json:"ams"`
	} `json:"kafka"`
	ServiceRegistry struct {
		Ams AmsConfig `json:"ams"`
	} `json:"serviceRegistry"`
}

// AmsConfig is a struct that contains the AMS configuration
type AmsConfig struct {
	TermsAndConditionsEventCode string `json:"termsAndConditionsEventCode"`
	TermsAndConditionsSiteCode  string `json:"termsAndConditionsSiteCode"`
	InstanceQuotaID             string `json:"instanceQuotaId"`
	TrialQuotaID                string `json:"trialQuotaId"`
}
