package hueclient

import (
	"fmt"
	"net/http"
)

func (c *Client) GetAllLights() (*LightList, error) {
	var lights LightList
	err := c.doRequest("clip/v2/resource/light", http.MethodGet, nil, &lights)
	if err != nil {
		return nil, err
	}
	return &lights, nil
}

func (c *Client) GetOneLightById(id string) (*LightListItem, error) {
	var lights LightList
	err := c.doRequest("clip/v2/resource/light/"+id, http.MethodGet, nil, &lights)
	if err != nil {
		return nil, err
	}

	if len(lights.Errors) > 0 {
		return nil, fmt.Errorf("failed to fetch light by id = %q due to: %s", id, lights.Errors[0].Description)
	}

	if len(lights.Data) == 0 {
		return nil, nil
	}
	return &lights.Data[0], nil
}

func (c *Client) UpdateOneLightById(id string, lightUpdate *LightBodyUpdate) (*ResourceIdentifier, error) {
	var lightUpdateResp LightUpdateResponse
	err := c.doRequest("clip/v2/resource/light/"+id, http.MethodPut, lightUpdate, &lightUpdateResp)
	if err != nil {
		return nil, fmt.Errorf("failed to update light by id = %q: %w", id, err)
	}

	if len(lightUpdateResp.Errors) > 0 {
		return nil, fmt.Errorf("failed to update light by id = %q due to: %s", id, lightUpdateResp.Errors[0].Description)
	}

	if len(lightUpdateResp.Data) == 0 {
		return nil, nil
	}

	return &lightUpdateResp.Data[0], nil
}

func (c *Client) TurnOnLightById(id string) error {
	lightUpdate := &LightBodyUpdate{
		On: &LightOnState{
			On: true,
		},
	}
	_, err := c.UpdateOneLightById(id, lightUpdate)
	return err
}

func (c *Client) TurnOffLightById(id string) error {
	lightUpdate := &LightBodyUpdate{
		On: &LightOnState{
			On: false,
		},
	}
	_, err := c.UpdateOneLightById(id, lightUpdate)
	return err
}
