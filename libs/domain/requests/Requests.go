package requests

import (
	"context"
	"encoding/json"
	"fmt"
	"golang_project/libs/domain"
	"golang_project/libs/domain/auth"
	"golang_project/libs/infra/util"
	"log"
	"reflect"

	"github.com/invopop/jsonschema"
	"github.com/pkg/errors"
)

type empty struct{}

var currentPkgPath string = reflect.TypeOf(empty{}).PkgPath()

// Request :
type Request interface {
	// Validate : 요청파라미터 검사
	Validate() error
	// RequestMethodType : 요청타입 - 조회/변경
	RequestMethodType() RequestMethodType

	IsAllowedRequesterType(requesterType auth.RequesterType) bool
}

// WriteRequest ...
type WriteRequest interface {
	Request
	// TargetResourceOwner : 대상 자원 소유자?
	TargetResourceOwner() string
}

// WriteResponse :
type WriteResponse struct {
	CommandID string
	Error     string
}

// WriteResponses :
type WriteResponses []WriteResponse

// CreateReadResponseInstanceFunc :
type CreateReadResponseInstanceFunc func() interface{}

// ReadRequest :
type ReadRequest interface {
	Request
	// ProcessReadRequest :
	ProcessReadRequest(ctx context.Context, svcName string) (response interface{}, err error)

	// GetCreateReadResponseInstanceFunc : 빈 응답 객체생성 대리자 반환
	GetCreateReadResponseInstanceFunc() CreateReadResponseInstanceFunc
}

// CreateRequestInstanceFunc :
type CreateRequestInstanceFunc func() Request

// RequestMapItem :
type RequestMapItem struct {
	RequestType        RequestMethodType
	IsCommon           bool
	CreateInstanceFunc CreateRequestInstanceFunc
	MethodInfo         domain.MethodInfo
}

// Print :
func (t RequestMapItem) Print() {
	log.Printf("\tIsCommon: %v, RequestMethodType: %s, Description: %s, ReqSchema: %s, RespSchema: %s\n",
		t.IsCommon, string(t.RequestType), t.MethodInfo.Description, t.MethodInfo.RequestSchema, t.MethodInfo.ResponseSchema)
	// log.Printf("\tRequestMethodType: %s\n", string(t.RequestMethodType))
	// log.Printf("\tDescription: %s\n", t.Description)
}

// RequestNameDescriptionMap : 요청명 | 설명 맵
type RequestNameDescriptionMap map[string]string

// RequestMap : Request Name/RequestMapItem
type RequestMap map[string]*RequestMapItem

// register : 성공시 request 이름 반환
func (t *RequestMap) register(emptyInstance Request, description string) string {
	if nil == emptyInstance {
		panic(errors.New("emptyInstance가 nil입니다"))
	}
	objType := reflect.TypeOf(emptyInstance).Elem()
	reqName := util.GetTypeName(emptyInstance)
	pkgPath := objType.PkgPath()
	isCommon := pkgPath == currentPkgPath
	// log.Printf(">> request.Register: %s.%s 요청 등록 시도 중...\n", pkgPath, reqName)
	createInstanceFunc := func() Request {
		return reflect.New(objType).Interface().(Request)
	}
	var err error
	var emptyRespInstance interface{}
	reqType := emptyInstance.RequestMethodType()
	if _, ok := emptyInstance.(ReadRequest); ok {
		if RequestMethodRead != reqType {
			panic(
				errors.Errorf(
					"ReadRequest 객체의 타입이 잘못 구현되어 있습니다-요청:%s, reqType:%s, ok: %v",
					reqName,
					reqType,
					ok))
		}
		emptyRespInstance = emptyInstance.(ReadRequest).GetCreateReadResponseInstanceFunc()()

	} else if _, ok := emptyInstance.(WriteRequest); ok {
		if RequestMethodWrite != reqType {
			panic(
				errors.Errorf(
					"RequestMethodWrite 객체의 타입이 잘못 구현되어 있습니다-요청:%s, reqType:%s",
					reqName,
					reqType))
		}
		emptyRespInstance = &WriteResponse{}
	}

	reqSchemaBytes, err := json.Marshal(jsonschema.Reflect(emptyInstance))
	if nil != err {
		panic(
			errors.Errorf(
				"Request 객체의 타입에서 요청 스키마를 추출하지 못했습니다-요청:%s, reqType:%s",
				reqName,
				reqType))
	}

	responseSchemaBytes, err := json.Marshal(jsonschema.Reflect(emptyRespInstance))
	if nil != err {
		panic(
			errors.Errorf(
				"Request 객체의 타입에서 응답 스키마를 추출하지 못했습니다-요청:%s, reqType:%s",
				reqName,
				reqType))
	}
	var allowedRequesters []string
	for _, requesterType := range auth.RequesterTypeList {
		if !emptyInstance.IsAllowedRequesterType(requesterType) {
			continue
		}
		allowedRequesters = append(allowedRequesters, string(requesterType))
	}

	item := &RequestMapItem{
		IsCommon:           isCommon,
		RequestType:        reqType,
		CreateInstanceFunc: createInstanceFunc,
		MethodInfo: domain.MethodInfo{
			Description:       description,
			RequestSchema:     string(reqSchemaBytes),
			ResponseSchema:    string(responseSchemaBytes),
			AllowedRequesters: allowedRequesters,
		},
	}
	(*t)[reqName] = item
	return reqName
}

// GetCommonRequestMap ...
func (t RequestMap) retrieveCommonRequestMap() RequestMap {
	commonMap := RequestMap{}
	for k, v := range t {
		if !v.IsCommon {
			continue
		}
		commonMap[k] = v
	}
	return commonMap
}

// GetLocalRequestMap ...
func (t *RequestMap) retrieveLocalRequestMap() RequestMap {
	localMap := RequestMap{}
	for k, v := range *t {
		if v.IsCommon {
			continue
		}
		localMap[k] = v
	}
	return localMap
}

// ToRequestNameMethodInfoMap : 요청명/설명맵으로 변환하여 반환
func (t RequestMap) ToRequestNameMethodInfoMap() map[string]domain.MethodInfo {
	reqNameMethodInfoMap := map[string]domain.MethodInfo{}
	for k, v := range t {
		reqNameMethodInfoMap[k] = v.MethodInfo
	}
	return reqNameMethodInfoMap
}

// Print :
func (t *RequestMap) Print(titleFormat string, a ...interface{}) {
	log.Println(fmt.Sprintf(titleFormat, a...))
	for k, v := range *t {
		log.Printf(" - Request Name: %s\n", k)
		if nil == v {
			continue
		}
		v.Print()
	}
}

func init() {
	log.Println(">>>>>>>>>>>>>>>> reqeusts package init <<<<<<<<<<<<<<<<<")
	log.Printf("requests(%s) package initialized.\n", currentPkgPath)
}
