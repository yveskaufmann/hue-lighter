package hueclient

type ResourceIdentifier struct {
	Action struct {
		Identity string `json:"identity,omitempty"`
	} `json:"action,omitempty"`
	// The duration in seconds to perform the identity action
	Duration *int `json:"duration,omitempty"`
}
