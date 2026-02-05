package hueclient

// Owner of the service, in case the owner service is deleted, the service also gets deleted
type DeviceOwner struct {
	// The unique id of the referenced resource
	RID string `json:"rid,omitempty"`

	// The type of the referenced resource
	RType ReferenceType `json:"rtype,omitempty"`
}

// ReferenceType defines the type of the referenced resource
type ReferenceType string

const (
	ReferenceTypeDevice                  ReferenceType = "device"
	ReferenceTypeBridgeHome              ReferenceType = "bridge_home"
	ReferenceTypeRoom                    ReferenceType = "room"
	ReferenceTypeZone                    ReferenceType = "zone"
	ReferenceTypeServiceGroup            ReferenceType = "service_group"
	ReferenceTypeLight                   ReferenceType = "light"
	ReferenceTypeButton                  ReferenceType = "button"
	ReferenceTypeBellButton              ReferenceType = "bell_button"
	ReferenceTypeRelativeRotary          ReferenceType = "relative_rotary"
	ReferenceTypeTemperature             ReferenceType = "temperature"
	ReferenceTypeLightLevel              ReferenceType = "light_level"
	ReferenceTypeMotion                  ReferenceType = "motion"
	ReferenceTypeCameraMotion            ReferenceType = "camera_motion"
	ReferenceTypeEntertainment           ReferenceType = "entertainment"
	ReferenceTypeContact                 ReferenceType = "contact"
	ReferenceTypeTamper                  ReferenceType = "tamper"
	ReferenceTypeConvenienceAreaMotion   ReferenceType = "convenience_area_motion"
	ReferenceTypeSecurityAreaMotion      ReferenceType = "security_area_motion"
	ReferenceTypeSpeaker                 ReferenceType = "speaker"
	ReferenceTypeGroupedLight            ReferenceType = "grouped_light"
	ReferenceTypeGroupedMotion           ReferenceType = "grouped_motion"
	ReferenceTypeGroupedLightLevel       ReferenceType = "grouped_light_level"
	ReferenceTypeDevicePower             ReferenceType = "device_power"
	ReferenceTypeDeviceSoftwareUpdate    ReferenceType = "device_software_update"
	ReferenceTypeZigbeeConnectivity      ReferenceType = "zigbee_connectivity"
	ReferenceTypeZgpConnectivity         ReferenceType = "zgp_connectivity"
	ReferenceTypeBridge                  ReferenceType = "bridge"
	ReferenceTypeMotionAreaCandidate     ReferenceType = "motion_area_candidate"
	ReferenceTypeWifiConnectivity        ReferenceType = "wifi_connectivity"
	ReferenceTypeZigbeeDeviceDiscovery   ReferenceType = "zigbee_device_discovery"
	ReferenceTypeHomekit                 ReferenceType = "homekit"
	ReferenceTypeMatter                  ReferenceType = "matter"
	ReferenceTypeMatterFabric            ReferenceType = "matter_fabric"
	ReferenceTypeScene                   ReferenceType = "scene"
	ReferenceTypeEntertainmentConfig     ReferenceType = "entertainment_configuration"
	ReferenceTypePublicImage             ReferenceType = "public_image"
	ReferenceTypeAuthV1                  ReferenceType = "auth_v1"
	ReferenceTypeBehaviorScript          ReferenceType = "behavior_script"
	ReferenceTypeBehaviorInstance        ReferenceType = "behavior_instance"
	ReferenceTypeGeofenceClient          ReferenceType = "geofence_client"
	ReferenceTypeGeolocation             ReferenceType = "geolocation"
	ReferenceTypeSmartScene              ReferenceType = "smart_scene"
	ReferenceTypeMotionAreaConfiguration ReferenceType = "motion_area_configuration"
	ReferenceTypeClip                    ReferenceType = "clip"
)
