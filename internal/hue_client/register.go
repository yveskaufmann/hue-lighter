package hueclient

import "fmt"

const APP_NAME = "hue-lighter"

type DeviceRegistrationRequest struct {
	DeviceType        string `json:"devicetype"`
	GenerateClientKey *bool  `json:"generateclientkey"`
}

type DeviceRegistrationResponse struct {
	Success *struct {
		Username  string `json:"username,omitempty"`
		ClientKey string `json:"clientkey,omitempty"`
	} `json:"success,omitempty"`

	Error *struct {
		Type        int    `json:"type,omitempty"`
		Address     string `json:"address,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"error,omitempty"`
}

func (r *DeviceRegistrationResponse) HasError() bool {
	return r.Error != nil
}

func (r *DeviceRegistrationResponse) ToError() error {
	if r.Error == nil {
		return nil
	}
	return fmt.Errorf("type %d: %s", r.Error.Type, r.Error.Description)
}

func (c *Client) RegisterDevice(name string) (*DeviceRegistrationResponse, error) {
	reqBody := DeviceRegistrationRequest{
		DeviceType:        FormatDeviceType(name),
		GenerateClientKey: &[]bool{true}[0],
	}

	var resp []DeviceRegistrationResponse
	err := c.doRequest("/api", "POST", reqBody, &resp)

	if err != nil {
		return nil, fmt.Errorf("failed to register device: %w", err)
	}

	return &resp[0], nil
}

func FormatDeviceType(name string) string {
	return fmt.Sprintf("%s#%s", APP_NAME, name)
}
