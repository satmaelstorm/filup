package logsEngine

type CounterOpts struct {
	Namespace string `json:"namespace,omitempty"`
	Name      string `json:"name,omitempty"`
	Subsystem string `json:"subsystem,omitempty"`
}
