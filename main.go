package grunway

const (
	BUILD_NUMBER   = 1
	VERSION_STRING = "0.1"
)

func Build() int64 {
	return BUILD_NUMBER
}
func Version() string {
	return VERSION_STRING
}
