package core

// ToolDefinition represents a single tool from commands.json
type ToolDefinition struct {
	ID              string        `json:"id"`
	Name            string        `json:"name"`
	Namespace       string        `json:"namespace"`
	Description     string        `json:"description"`
	Author          interface{}   `json:"author,omitempty"`
	Version         string        `json:"version"`
	Capabilities    []string      `json:"capabilities,omitempty"`
	Platforms       []string      `json:"platforms,omitempty"`
	Architectures   []string      `json:"architectures,omitempty"`
	RiskLevel       string        `json:"risk_level,omitempty"`
	ExecutionPolicy string        `json:"execution_policy,omitempty"`
	TrustLevel      string        `json:"trust_level,omitempty"`
	Features        []string      `json:"features,omitempty"`
	Techniques      []string      `json:"techniques,omitempty"`
	Parameters      []Parameter   `json:"parameters,omitempty"`
	Execution       Execution     `json:"execution"`
	Install         []Install     `json:"install,omitempty"`
	Phase           string        `json:"phase,omitempty"`
	MitreIDs        []string      `json:"mitre_ids,omitempty"`
	Dependencies    []string      `json:"dependencies,omitempty"`
	RelatedTools    []string      `json:"related_tools,omitempty"`
	GlobalVars      map[string]string `json:"global_vars,omitempty"`
}

type Parameter struct {
	Name         string      `json:"name"`
	TemplateKey  string      `json:"template_key,omitempty"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Description  string      `json:"description"`
	Aliases      []string    `json:"aliases,omitempty"`
	Enum         []string    `json:"enum,omitempty"`
	Pattern      string      `json:"pattern,omitempty"`
	Minimum      *float64    `json:"minimum,omitempty"`
	Maximum      *float64    `json:"maximum,omitempty"`
}

type Execution struct {
	Template       string            `json:"template"`
	Sandbox        string            `json:"sandbox"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty"`
	Shell          bool              `json:"shell,omitempty"`
	Workdir        string            `json:"workdir,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Container      *Container        `json:"container,omitempty"`
}

type Container struct {
	Image string `json:"image"`
}

type Install struct {
	Method      string   `json:"method"`
	PackageName string   `json:"package_name,omitempty"`
	RepoURL     string   `json:"repo_url,omitempty"`
	Commands    []string `json:"commands"`
}

// InstallStatus tracks whether a tool is installed and how
type InstallStatus struct {
	ToolName       string
	OnPath         bool
	PathLocation   string
	PackageManager string
	Version        string
	DockerImage    bool
}

// ToolSource represents from where we load tool data
type ToolSource string

const (
	SourceCache ToolSource = "cache"
	SourceSite  ToolSource = "site"
	SourceFile  ToolSource = "file"
)
