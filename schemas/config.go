package schemas

type SiteConfig struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Theme            string            `json:"theme"` // e.g., "basic"
	Colors           Colors            `json:"colors"`
	DefaultThemeMode string            `json:"defaultThemeMode"` // e.g., "light", "dark"
	Logo             Logo              `json:"logo"`
	Favicon          string            `json:"favicon"` // Path to the favicon image
	Background       Background        `json:"background"`
	Banner           Banner            `json:"banner"`
	APIServer        *string           `json:"apiServer,omitempty"`
	Navbar           []NavbarComponent `json:"navbar"`
	Navigation       *Navigation       `json:"navigation"`
	Footer           []SocialLink      `json:"footer"`
	TryButtonBaseUrl     *string           `json:"try_button_base_url"`
}

type Background struct {
	Colors Colors `json:"colors"`
	Image  Image  `json:"image"`
}

type Colors struct {
	Dark  string `json:"dark"`
	Light string `json:"light"`
}

type Logo struct {
	Light     string `json:"light"`     // Paths to logo images for light theme
	Dark      string `json:"dark"`      // Paths to logo images for dark theme
	Reference string `json:"reference"` // Link to redirect when logo is clicked
}

type Favicon struct {
	Light string `json:"light"` // Path to the favicon image for light theme
	Dark  string `json:"dark"`  // Path to the favicon image for dark theme
}

type Image struct {
	Light string `json:"light"` // Path to the image for light theme
	Dark  string `json:"dark"`  // Path to the image for dark theme
}

type Banner struct {
	Message string `json:"message"` // Message to display in the banner
	Colors  Colors `json:"colors"`  // Colors for the banner
}

type NavbarComponent struct {
	Type      string `json:"type"`      // e.g., "link", "button"
	Label     string `json:"label"`     // Text displayed on the component
	Reference string `json:"reference"` // Link or action associated with the component
}

type SocialLink struct {
	Type      string `json:"type"`      // e.g., "github", "linkedin", "x", "facebook", "youtube", "link", "instagram"
	Reference string `json:"reference"` // Link to the social media profile
}

type Navigation struct {
	Versions []Version `json:"versions,omitempty"` // List of versions for the navigation. Optional
	Tabs     []Tab     `json:"tabs,omitempty"`     // List of tabs for the navigation. Optional if there are no versions
	Items    []Item    `json:"items,omitempty"`    // List of items in the navigation. Optional if there are no versions or tabs
}

type Version struct {
	Version string `json:"version"`         // Version identifier, e.g., "v1", "v2"
	Tabs    []Tab  `json:"tabs,omitempty"`  // List of tabs associated with the version. Optional
	Items   []Item `json:"items,omitempty"` // List of items in the version. Optional if there are no tabs
}

type Tab struct {
	Tab   string `json:"tab"`   // Tab identifier, e.g., "latest", "stable"
	Items []Item `json:"items"` // List of items associated with the tab.
}

type Item struct {
	ApiReference
	Type     string `json:"type"`               // Type of the item, e.g., "group", "dropdown", "page", "openapi", "swagger"
	Title    string `json:"title"`              // Title of the item, e.g., "Documentation", "API"
	Page     string `json:"page"`               // Path to the page or resource, e.g., "docs/introduction.md". Only used if Type is "page"
	ApiRefMethod *string `json:"method,omitempty"` //Optional for api ref pages
	ApiRefPath *string `json:"api_ref_path,omitempty"` //Optional path of the openapi file
	ApiRefRequest *string `json:"request,omitempty"`
	Children []Item `json:"children,omitempty"` // Optional list of child items for nested navigation. Only used if Type is "group" or "dropdown"
}

type ApiReference struct {
	Method    string `json:"method"`    // HTTP method, e.g., "GET", "POST"
	Reference string `json:"reference"` // API endpoint or URL
	Path      string `json:"path"`      // Path to the API documentation or resource
}
