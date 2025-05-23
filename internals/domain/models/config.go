package config

type Tab struct {
	Group string   `json:"group"`
	Pages []string `json:"pages"`
}

type LogoModel struct {
	Light string `json:"light"`
	Dark  string `json:"dark"`
}
type ThemeValue struct {
	Dark  any `json:"dark"`
	Light any `json:"light"`
}

type Color struct {
	Hue        ThemeValue `json:"hue"`
	Saturation ThemeValue `json:"saturation"`
	Lightness  ThemeValue `json:"lightness"`
}

type ColorScheme struct {
	Color           Color      `json:"color"`
	BackgroundColor ThemeValue `json:"backgroundColor"`
}

type Version struct {
	MainView string      `json:"index"`
	Tabs     []Tab       `json:"tabs"`
	Theme    string      `json:"theme,omitempty"`
	Banner   string      `json:"banner,omitempty"`
	Github   string      `json:"github,omitempty"`
	Logo     LogoModel   `json:"logo"`
	Name     string      `json:"name"`
	Colors   ColorScheme `json:"colors"`
}

type WebConfig struct {
	DefaultVersion string             `json:"defaultVersion"`
	Versions       map[string]Version `json:"versions"`
}
