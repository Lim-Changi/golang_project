package requests

// RequestMethodType : 요청타입
type RequestMethodType string

// RequestMethodTypes
const (
	RequestMethodRead  RequestMethodType = "read"
	RequestMethodWrite RequestMethodType = "write"
)

// IsValidRequestMethodType :
func IsValidRequestMethodType(methodType RequestMethodType) bool {
	switch methodType {
	case RequestMethodRead:
		fallthrough
	case RequestMethodWrite:
		return true
	default:
		return false
	}
}
