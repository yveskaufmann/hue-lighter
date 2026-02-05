package hueclient

type LightFunction string

const (
	FunctionFunctional LightFunction = "functional"
	FunctionDecorative LightFunction = "decorative"
	FunctionMixed      LightFunction = "mixed"
	FunctionUnknown    LightFunction = "unknown"
)

type LightMeta struct {
	LightProductData
	FixedMired int    `json:"fixed_mired,omitempty"`
	Name       string `json:"name,omitempty"`
}

type LightProductData struct {
	Name       string         `json:"name,omitempty"`
	Archetype  LightArchetype `json:"archetype,omitempty"`
	FixedMired int            `json:"fixed_mired,omitempty"`
}

type LightOnState struct {
	On bool `json:"on"`
}

type LightDimmingState struct {
	Dimming     float32 `json:"dimming,omitempty"`
	MinDimLevel float32 `json:"min_dim_level,omitempty"`
}

type LightDimmingDeltaState struct {
	Action          string   `json:"action,omitempty"`
	BrightnessDelta *float64 `json:"brightness_delta,omitempty"`
}

type LightColorTemperature struct {
	Mirek *int `json:"mirek,omitempty"`
}

type ColorTemperatureAction string

const (
	ColorTemperatureActionUp   ColorTemperatureAction = "up"
	ColorTemperatureActionDown ColorTemperatureAction = "down"
	ColorTemperatureActionStop ColorTemperatureAction = "stop"
)

type LightColorTemperatureDelta struct {
	Action     ColorTemperatureAction `json:"action,omitempty"`
	MirekDelta *int                   `json:"mirek_delta,omitempty"`
}

type LightColor struct {
	// CIE XY gamut position
	XY *struct {
		X float32 `json:"x,omitempty"`
		Y float32 `json:"y,omitempty"`
	} `json:"xy,omitempty"`
}

type Dynamics struct {
	Duration *int `json:"duration,omitempty"`
	Speed    *int `json:"speed,omitempty"`
}

type Alert struct {
	// Alert effect
	Action string `json:"action,omitempty"`
}

type SignalType string

const (
	SignalTypeIdentify SignalType = "identify"
)

const (
	SignalTypeNoSignal    SignalType = "no_signal"
	SignalTypeOnOff       SignalType = "on_off"
	SignalTypeOnOffColor  SignalType = "on_off_color"
	SignalTypeAlternating SignalType = "alternating"
)

type Signaling struct {
	// Signal to set the light to
	Signal SignalType `json:"signal,omitempty"`

	// Duration in milliseconds for the signaling effect
	Duration int `json:"duration,omitempty"`

	// List of colors to apply to the signal (not supported by all signals)
	Colors []LightColor `json:"colors,omitempty"`
}

type GradientMode string

const (
	GradientModeInterpolatedPalette       GradientMode = "interpolated_palette"
	GradientModeInterpolatedPaletteMirror GradientMode = "interpolated_palette_mirrored"
	GradientModeRandomPixelated           GradientMode = "random_pixelated"
	GradientModeSegmentedPalette          GradientMode = "segmented_palette"
)

type Gradient struct {
	Points []struct {
		Color LightColor `json:"color,omitempty"`
	} `json:"points,omitempty"`
	Mode *GradientMode `json:"mode,omitempty"`
}

type EffectType string

const (
	EffectPrism      EffectType = "prism"
	EffectOpal       EffectType = "opal"
	EffectGlisten    EffectType = "glisten"
	EffectSparkle    EffectType = "sparkle"
	EffectFire       EffectType = "fire"
	EffectCandle     EffectType = "candle"
	EffectUnderwater EffectType = "underwater"
	EffectCosmos     EffectType = "cosmos"
	EffectSunbeam    EffectType = "sunbeam"
	EffectEnchant    EffectType = "enchant"
	EffectNoEffect   EffectType = "no_effect"
)

type TimedEffectType string

const (
	TimedEffectSunrise  TimedEffectType = "sunrise"
	TimedEffectSunset   TimedEffectType = "sunset"
	TimedEffectNoEffect TimedEffectType = "no_effect"
)

type PowerupPreset string

const (
	PowerupPresetSafety      PowerupPreset = "safety"
	PowerupPresetPowerfail   PowerupPreset = "powerfail"
	PowerupPresetLastOnState PowerupPreset = "last_on_state"
	PowerupPresetCustom      PowerupPreset = "custom"
)

type PowerupOnMode string

const (
	PowerupOnModeOn       PowerupOnMode = "on"
	PowerupOnModeToggle   PowerupOnMode = "toggle"
	PowerupOnModePrevious PowerupOnMode = "previous"
)

type PowerupDimmingMode string

const (
	PowerupDimmingModeDimming  PowerupDimmingMode = "dimming"
	PowerupDimmingModePrevious PowerupDimmingMode = "previous"
)

type PowerupColorMode string

const (
	PowerupColorModeColorTemperature PowerupColorMode = "color_temperature"
	PowerupColorModeColor            PowerupColorMode = "color"
	PowerupColorModePrevious         PowerupColorMode = "previous"
)

type ContentOrientation string

const (
	ContentOrientationHorizontal ContentOrientation = "horizontal"
	ContentOrientationVertical   ContentOrientation = "vertical"
)

type ContentOrder string

const (
	ContentOrderForward  ContentOrder = "forward"
	ContentOrderReversed ContentOrder = "reversed"
)

// Nested types for effects_v2
type EffectsV2 struct {
	Action *EffectAction `json:"action,omitempty"`
}

type EffectAction struct {
	Effect     EffectType        `json:"effect"`
	Parameters *EffectParameters `json:"parameters,omitempty"`
	Speed      *float64          `json:"speed,omitempty"` // 0..1
}

type EffectParameters struct {
	Color *LightColor `json:"color,omitempty"`
}

// Timed effects
type TimedEffects struct {
	Effect   TimedEffectType `json:"effect"`
	Duration *int            `json:"duration,omitempty"` // ms
}

// Powerup configuration
type Powerup struct {
	Preset    PowerupPreset     `json:"preset"`
	On        *PowerupOn        `json:"on,omitempty"`
	Dimming   *PowerupDimming   `json:"dimming,omitempty"`
	Color     *PowerupColor     `json:"color,omitempty"`
	ColorTemp *PowerupColorTemp `json:"color_temperature,omitempty"`
}

type PowerupOn struct {
	Mode PowerupOnMode `json:"mode"`
	On   *struct {
		On bool `json:"on"`
	} `json:"on,omitempty"`
}

type PowerupDimming struct {
	Mode    PowerupDimmingMode `json:"mode"`
	Dimming *struct {
		Brightness float64 `json:"brightness"`
	} `json:"dimming,omitempty"`
}

type PowerupColor struct {
	Mode  PowerupColorMode `json:"mode"`
	Color *LightColor      `json:"color,omitempty"`
}

type PowerupColorTemp struct {
	Mirek *int `json:"mirek,omitempty"`
}

// Content configuration
type ContentConfiguration struct {
	Orientation *ContentOrientationConfig `json:"orientation,omitempty"`
	Order       *ContentOrderConfig       `json:"order,omitempty"`
}

type ContentOrientationConfig struct {
	Orientation ContentOrientation `json:"orientation"`
}

type ContentOrderConfig struct {
	Order ContentOrder `json:"order"`
}

type LightListItem struct {
	ID    string      `json:"id,omitempty"`
	IDV1  string      `json:"id_v1,omitempty"`
	Owner DeviceOwner `json:"owner"`
	Type  string      `json:"type,omitempty"`

	Meta         LightMeta               `json:"metadata,omitempty"`
	ProductData  LightProductData        `json:"product_data,omitempty"`
	Identity     interface{}             `json:"identity,omitempty"`
	ServiceId    int                     `json:"service_id,omitempty"`
	On           LightOnState            `json:"on,omitempty"`
	Dimming      *LightDimmingState      `json:"dimming,omitempty"`
	DimmingDelta *LightDimmingDeltaState `json:"dimming_delta,omitempty"`
}

type LightBodyUpdate struct {
	Type                  string                      `json:"type,omitempty"`
	Meta                  *LightMeta                  `json:"metadata,omitempty"`
	Identity              *ResourceIdentifier         `json:"identity,omitempty"`
	On                    *LightOnState               `json:"on,omitempty"`
	Dimming               *LightDimmingState          `json:"dimming,omitempty"`
	DimmingDelta          *LightDimmingDeltaState     `json:"dimming_delta,omitempty"`
	ColorTemperature      *LightColorTemperature      `json:"color_temperature,omitempty"`
	ColorTemperatureDelta *LightColorTemperatureDelta `json:"color_temperature_delta,omitempty"`
	Color                 *LightColor                 `json:"color,omitempty"`
	Dynamics              *Dynamics                   `json:"dynamics,omitempty"`
	Alert                 *Alert                      `json:"alert,omitempty"`
	Signaling             *Signaling                  `json:"signaling,omitempty"`
	Gradient              *Gradient                   `json:"gradient,omitempty"`
	EffectsV2             *EffectsV2                  `json:"effects_v2,omitempty"`
	TimedEffects          *TimedEffects               `json:"timed_effects,omitempty"`
	Powerup               *Powerup                    `json:"powerup,omitempty"`
	ContentConfiguration  *ContentConfiguration       `json:"content_configuration,omitempty"`
}

type LightUpdateResponse struct {
	Data   []ResourceIdentifier `json:"data,omitempty"`
	Errors []struct {
		Description string `json:"description,omitempty"`
	} `json:"errors,omitempty"`
}

type LightList struct {
	Data   []LightListItem `json:"data,omitempty"`
	Errors []struct {
		Description string `json:"description,omitempty"`
	} `json:"errors,omitempty"`
}
