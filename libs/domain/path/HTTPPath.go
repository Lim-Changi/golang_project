package path

import (
	"golang_project/libs/domain/auth"
	"golang_project/libs/domain/requests"
	"strings"

	"github.com/pkg/errors"
)

const (
	// HTTPRequestPathElementCount : "서비스/[admin/user/sys]/[read,write]/XXXXXXRequest"를 파싱하면 4개의 요소가 나와야 한다
	HTTPRequestPathElementCount = 4
)

// HTTPRequestPath : 파싱된 HTTP 요청 경로 정보
type HTTPRequestPath struct {
	ServiceName       string                     `jsonschema:"description=서비스명,example=mms"`
	AuthType          auth.RequestAuthType       `jsonschema:"description=인증타입,enum=user,enum=iam"`
	RequestMethodType requests.RequestMethodType `jsonschema:"description=요청타입,enum=read,enum=write"`
	RequesterType     auth.RequesterType         `jsonschema:"description=요청자타입,enum=admin,enum=user"`
	RequestName       string                     `jsonschema:"description=요청명"`
}

// ParseHTTPRequestPath : HTTP 요청 경로를 파싱
// HTTP 경로형식 /서비스/[admin/user]/[read,write]/[요청명]
// API Gateway 설정시 아래 형식으로 설정한다.
// /서비스/[admin/user]/[read,write]/{+proxy}
func ParseHTTPRequestPath(path string) (*HTTPRequestPath, error) {
	if path = strings.TrimSpace(path); "" == path {
		return nil, errors.New("path is empty")
	}
	// 주소는 "/"로 시작해야 한다.
	if !strings.HasPrefix(path, "/") {
		return nil, errors.Errorf("HTTP path is not start with '/' - %s", path)
	}
	tokens := strings.Split(strings.TrimPrefix(path, "/"), "/")
	// "[svc]/[iam,user]/[read,write]/XXXXXXRequest"를 파싱하면 4개의 요소가 나와야 한다.
	if len(tokens) < HTTPRequestPathElementCount {
		return nil, errors.Errorf("HTTP path must consist of %d elements - path: %s", HTTPRequestPathElementCount, path)
	}

	requester := auth.RequesterType(tokens[1])
	if !auth.IsValidRequesterType(requester) {
		return nil, errors.Errorf("%s is invalid requester type", requester)
	}

	rp := HTTPRequestPath{
		ServiceName:   tokens[0],
		RequesterType: requester,
		RequestName:   tokens[3],
	}
	if "" == rp.ServiceName {
		return nil, errors.New("ServiceName is Empty")
	} else if "" == rp.RequestName {
		return nil, errors.New("RequestName is Empty")
	}
	rp.RequestMethodType = requests.RequestMethodType(tokens[2])
	if !requests.IsValidRequestMethodType(rp.RequestMethodType) {
		return nil, errors.Errorf("Unknown Request Type: %s", rp.RequestMethodType)
	}

	// AuthType
	rp.AuthType = auth.RequesterType2AuthType(rp.RequesterType)
	if !auth.IsValidRequestAuthType(rp.AuthType) {
		return nil, errors.Errorf(
			"Invalid auth type: %s", rp.AuthType)
	}
	return &rp, nil
}
