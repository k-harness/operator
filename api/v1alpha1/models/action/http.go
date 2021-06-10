package action

//+kubebuilder:object:generate=true

type HTTP struct {
	Addr string `json:"addr"`
	//+kubebuilder:validation:Enum=GET;POST;PUT;DELETE
	Method string            `json:"method"`
	Path   *string           `json:"path,omitempty"`
	Query  map[string]string `json:"query,omitempty"`
}
