package clients

// Page contains page metadata that helps navigation.
type Page struct {
	Total        uint64   `json:"total"`
	Offset       uint64   `json:"offset"`
	Limit        uint64   `json:"limit"`
	Name         string   `json:"name,omitempty"`
	Order        string   `json:"order,omitempty"`
	Dir          string   `json:"dir,omitempty"`
	Metadata     Metadata `json:"metadata,omitempty"`
	Disconnected bool     // Used for connected or disconnected lists
	Owner        string
	Tag          string
	SharedBy     string
	Status       Status
	Action       string
	Subject      string
	IDs          []string
	Identity     string
}
