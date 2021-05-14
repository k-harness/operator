package action

//+kubebuilder:object:generate=true

type HTTP struct {
	Addr   string  `json:"addr"`
	Method string  `json:"method"`
	Path   *string `json:"path,omitempty"`
	Query  *string `json:"query,omitempty"`
}
