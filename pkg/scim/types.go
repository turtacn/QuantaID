package scim

// Resource is the common struct for all SCIM resources
type Resource struct {
	Schemas []string `json:"schemas"`
	ID      string   `json:"id,omitempty"`
	Meta    *Meta    `json:"meta,omitempty"`
}

// Meta contains metadata about the resource
type Meta struct {
	ResourceType string `json:"resourceType"`
	Created      string `json:"created,omitempty"`
	LastModified string `json:"lastModified,omitempty"`
	Location     string `json:"location,omitempty"`
	Version      string `json:"version,omitempty"`
}

// User represents a SCIM User Resource
type User struct {
	Resource
	UserName          string   `json:"userName"`
	ExternalID        string   `json:"externalId,omitempty"`
	Active            bool     `json:"active"`
	Name              *Name    `json:"name,omitempty"`
	Emails            []Email  `json:"emails,omitempty"`
	PhoneNumbers      []Phone  `json:"phoneNumbers,omitempty"`
	Groups            []GroupRef `json:"groups,omitempty"`
	// Enterprise extension fields can be added here or via a map if dynamic handling is needed
	// For this task, we will stick to core fields + minimal extension support if needed
}

type Name struct {
	Formatted       string `json:"formatted,omitempty"`
	FamilyName      string `json:"familyName,omitempty"`
	GivenName       string `json:"givenName,omitempty"`
	MiddleName      string `json:"middleName,omitempty"`
	HonorificPrefix string `json:"honorificPrefix,omitempty"`
	HonorificSuffix string `json:"honorificSuffix,omitempty"`
}

type Email struct {
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type Phone struct {
	Value   string `json:"value,omitempty"`
	Type    string `json:"type,omitempty"`
	Primary bool   `json:"primary,omitempty"`
}

type GroupRef struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Ref     string `json:"$ref,omitempty"`
}

// Group represents a SCIM Group Resource
type Group struct {
	Resource
	DisplayName string   `json:"displayName"`
	ExternalID  string   `json:"externalId,omitempty"`
	Members     []Member `json:"members,omitempty"`
}

type Member struct {
	Value   string `json:"value"`
	Display string `json:"display,omitempty"`
	Ref     string `json:"$ref,omitempty"`
}

// ListResponse represents a SCIM List Response
type ListResponse struct {
	Schemas      []string      `json:"schemas"`
	TotalResults int           `json:"totalResults"`
	ItemsPerPage int           `json:"itemsPerPage,omitempty"`
	StartIndex   int           `json:"startIndex,omitempty"`
	Resources    []interface{} `json:"Resources"`
}

// Error represents a SCIM Error Response
type Error struct {
	Schemas []string `json:"schemas"`
	Status  string   `json:"status"`
	ScimType string  `json:"scimType,omitempty"`
	Detail  string   `json:"detail,omitempty"`
}
