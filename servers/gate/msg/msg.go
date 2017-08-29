package msg

var Processor *GateProcessor = NewProcessor()

func init() {
}

type Hello struct {
	Id   int
	Name string
}
