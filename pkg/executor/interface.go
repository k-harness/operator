package executor

type Request struct {
	Body   []byte
	Type   string
	Header map[string]string
}
