package read

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"golang_project/libs/domain/auth"
	"golang_project/libs/domain/requests"
	"golang_project/libs/infra/util"
	"log"
	"reflect"
	"strings"
	"time"
)

// RawRequest :
type RawRequest struct {
	Name  string          `jsonschema:"description=요청명"`
	Param json.RawMessage `jsonschema:"description=요청파라미터-JSON"`
}

// RequestUnmarshaler : RawRequest ==> ReadRequest
type RequestUnmarshaler interface {
	Unmarshal(rawReq *RawRequest) (requests.ReadRequest, error)
	// 서비스 공통
	IsCommandRequest(requestName string) bool
	// 각 서비스 별
	IsLocalRequest(requestName string) bool
}

// requestUnmarshaler :
type requestUnmarshaler struct {
	// lowercase string, func()
	commonReqNameInstanceFuncMap map[string]requests.CreateRequestInstanceFunc
	localReqNameInstanceFuncMap  map[string]requests.CreateRequestInstanceFunc
}

// InitRequestUnmarshaler :
// localReqNameInstanceFuncMap - request name : func() requests.Request { return &domain.GetSiteRequest{} }
func InitRequestUnmarshaler(readRequests *requests.ReadRequests) (RequestUnmarshaler, error) {
	if nil == readRequests {
		return nil, errors.Errorf("%s: readRequests가 입력되지 않았습니다", util.GetCurrentFuncName())
	}
	// localReqNameInstanceFuncMap의 키값을 모두 소문자로 바꾸고, 값인 함수가 nil 인지 체크
	refinedLocalReqNameInstanceFuncMap := map[string]requests.CreateRequestInstanceFunc{}
	for k, v := range readRequests.GetLocalRequestMap() {
		if nil == v {
			return nil, errors.Errorf("%s 요청 인스턴스 생성함수가 nil 입니다", k)
		}
		refinedLocalReqNameInstanceFuncMap[strings.ToLower(k)] = v.CreateInstanceFunc
	}

	// commonReqNameInstanceFuncMap의 키값을 모두 소문자로 바꾸고, 값인 함수가 nil 인지 체크
	refinedCommonReqNameInstanceFuncMap := map[string]requests.CreateRequestInstanceFunc{}
	for k, v := range readRequests.GetCommonRequestMap() {
		if nil == v {
			return nil, errors.Errorf("%s 요청 인스턴스 생성함수가 nil 입니다", k)
		}
		refinedCommonReqNameInstanceFuncMap[strings.ToLower(k)] = v.CreateInstanceFunc
	}

	return &requestUnmarshaler{
		commonReqNameInstanceFuncMap: refinedCommonReqNameInstanceFuncMap,
		localReqNameInstanceFuncMap:  refinedLocalReqNameInstanceFuncMap,
	}, nil
}

func (t *requestUnmarshaler) IsCommandRequest(requestName string) bool {
	_, ok := t.commonReqNameInstanceFuncMap[strings.ToLower(requestName)]
	return ok
}
func (t *requestUnmarshaler) IsLocalRequest(requestName string) bool {
	_, ok := t.localReqNameInstanceFuncMap[strings.ToLower(requestName)]
	return ok
}

// Unmarshal : RawRequest => ReadRequest
func (t *requestUnmarshaler) Unmarshal(rawReq *RawRequest) (requests.ReadRequest, error) {
	reqName := strings.ToLower(rawReq.Name)
	var getReqInstanceFunc requests.CreateRequestInstanceFunc
	var ok bool
	getReqInstanceFunc, ok = t.commonReqNameInstanceFuncMap[reqName]
	if !ok {
		getReqInstanceFunc, ok = t.localReqNameInstanceFuncMap[reqName]
		if !ok {
			return nil, errors.Errorf("지원하지 않는 요청 타입입니다 - %s", rawReq.Name)
		}
	}
	if nil == getReqInstanceFunc {
		return nil, errors.Errorf("%s 요청에 대한 세부 요청 인스턴스 생성 함수가 nil 입니다", rawReq.Name)
	}
	reqInstance := getReqInstanceFunc()
	if nil == reqInstance {
		return nil, errors.Errorf("%s 요청에 대한 세부 요청 인스턴스 생성 함수가 nil을 반환했습니다", rawReq.Name)
	}
	err := json.Unmarshal(rawReq.Param, reqInstance)
	if nil != err {
		return nil, errors.Wrapf(
			err, "%s 요청 파라미터를 복원하지 못했습니다 - 파라미터: %s, 대상객체 타입: <%s>",
			rawReq.Name,
			string(rawReq.Param),
			reflect.TypeOf(reqInstance))
	}
	return reqInstance.(requests.ReadRequest), nil
}

// Service :
type Service interface {
	Process(ctx context.Context, rawReq *RawRequest) (response interface{}, err error)
}

// Service : DetailRequestHandler implements
type service struct {
	RequestUnmarshaler
	svcName string
}

// InitService :
func InitService(
	svcName string,
	reqUnmarshaler RequestUnmarshaler) Service {
	return &service{
		RequestUnmarshaler: reqUnmarshaler,
		svcName:            svcName,
	}
}

// Process :
func (t *service) Process(ctx context.Context, rawReq *RawRequest) (interface{}, error) {
	if nil == rawReq {
		return nil, nil
	}
	if "" == rawReq.Name {
		log.Printf("readRequest.Name이 입력되지 않았습니다. Param:%s \n", string(rawReq.Param))
		return nil, nil
	}

	// for debug
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf("!!! ReadSvc elapsed: %s (request: %s)\n", elapsed, rawReq.Name)
	}()
	//

	// Unmarshal
	readRequestInstance, err := t.RequestUnmarshaler.Unmarshal(rawReq)
	if nil != err {
		return nil, err
	} else if nil == readRequestInstance {
		return nil, errors.Errorf("%s 요청을 복원하지 못했습니다", rawReq.Name)
	}

	reqCtx := auth.ExtractRequestContext(ctx)
	if nil == reqCtx {
		return nil, errors.New("RequestContext를 context에서 추출하지 못했습니다")
	}
	// 요청 경로와 허용 요청자 검사(admin, user, sys)- 예)admin 요청경로로 들어온 요청인데 일반 사용자가 호출할 수 없다.
	if !readRequestInstance.IsAllowedRequesterType(reqCtx.RequesterType) {
		return nil, errors.Wrapf(err, "%s 요청은 %s 타입 요청자를 허용하지 않습니다.", rawReq.Name, reqCtx.RequesterType)
	}

	// 요청파라미터 유효성 검사
	err = readRequestInstance.Validate()
	if nil != err {
		return nil, err
	}
	// for debug
	// log.Printf("ReadSvc unmarshaled readRequestInstance type: %s\n", util.GetTypeName(readRequestInstance))
	//
	return readRequestInstance.ProcessReadRequest(ctx, t.svcName)
}
