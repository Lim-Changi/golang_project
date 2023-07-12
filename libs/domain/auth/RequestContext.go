package auth

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	lambdaSvc "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/pkg/errors"
	"golang_project/libs/infra/util"
	"log"
)

// UserAuthRequestContext :
type UserAuthRequestContext struct {
	UserID string `json:",omitempty" jsonschema:"description= 유저 ID "`
	// TODO: 접근 가능한 자원 및 서비스를 제한할 경우 추가 객체 필요
}

// IAMAuthRequestContext :
type IAMAuthRequestContext struct {
	UserArn   string `json:",omitempty" jsonschema:"description=IAM UserARN"`
	AccessKey string `json:",omitempty" jsonschema:"description=IAM AccessKey"`
	User      string `json:",omitempty" jsonschema:"description=IAM User"`
}

// RequestContext : lambda authorizer 로 전달되는 context, 아직 내용 미확정
type RequestContext struct {
	// 인증 타입
	AuthType      RequestAuthType `jsonschema:"description=인증타입,enum=Lambda,enum=IAM"`
	RequesterType RequesterType   `jsonschema:"description=요청자 경로 타입,enum=admin,enum=user"`

	UserAuthRequestContext UserAuthRequestContext `json:",omitempty" jsonschema:"description=Lambda 인증일 경우 요청 컨텍스트"`
	IAMAuthRequestContext  IAMAuthRequestContext  `json:",omitempty" jsonschema:"description=IAM 인증일 경우 요청 컨텍스트"`
}

func (t RequestContext) Requester() string {
	switch t.RequesterType {
	case RequesterAdmin:
		return t.IAMAuthRequestContext.User
	case RequesterUser:
		return t.UserAuthRequestContext.UserID
	default:
		return ""
	}
}

// RequestContextKey : RequestContext 를 context 에 넣어 전달할때 태그 문자열
type requestContextKeyType string

const (
	RequestContextKey requestContextKeyType = "RequestContext"
)

// EmbedRequestContext : context 에 RequestContext 를 넣는다.
func EmbedRequestContext(ctx context.Context, reqCtx *RequestContext) context.Context {
	return context.WithValue(ctx, RequestContextKey, reqCtx)
}

// ExtractRequestContext : context 에서 RequestContext 를 추출한다.
func ExtractRequestContext(ctx context.Context) *RequestContext {
	if v := ctx.Value(RequestContextKey); nil != v {
		return v.(*RequestContext)
	}
	return nil
}

// InvokeGetAccessibleResourcesLambdaInput ...
type InvokeGetAccessibleResourcesLambdaInput struct {
	_            struct{}
	LambdaClient *lambdaSvc.Lambda
	LambdaName   string
	ServiceName  string //
	UserID       string
}

// Validate :
func (t InvokeGetAccessibleResourcesLambdaInput) Validate() error {
	if nil == t.LambdaClient {
		return errors.New("LambdaClient가 nil입니다")
	}
	if t.LambdaName == "" {
		return errors.New("LambdaName이 공백입니다")
	}
	if t.ServiceName == "" {
		return errors.New("ServiceName이 공백입니다")
	}
	if t.UserID == "" {
		return errors.New("UserID가 공백입니다")
	}
	return nil
}

// RetrieveAuthorizerRequestContextInput ...
type RetrieveAuthorizerRequestContextInput struct {
	_             struct{}
	AuthType      RequestAuthType
	RequesterType RequesterType
	Event         *events.APIGatewayProxyRequest
	ServiceName   string
}

// RetrieveAuthorizerRequestContext :
func RetrieveAuthorizerRequestContext(ctx context.Context, in *RetrieveAuthorizerRequestContextInput) (*RequestContext, error) {
	if nil == in {
		return nil, errors.Errorf("%s: 입력이 nil 입니다", util.GetCurrentFuncName())
	} else if nil == in.Event {
		return nil, errors.Errorf("%s: 입력이 nil 입니다", util.GetCurrentFuncName())
	} else if !IsValidRequesterType(in.RequesterType) {
		return nil, errors.Errorf("%s: RequesterType(%s)이 잘못입력되었습니다", util.GetCurrentFuncName(), in.RequesterType)
	}

	reqCtx := RequestContext{AuthType: in.AuthType, RequesterType: in.RequesterType}
	switch in.AuthType {
	case RequestAuthLambda:
		if len(in.Event.RequestContext.Authorizer) == 0 {
			bytes, _ := json.Marshal(in.Event)
			return nil, errors.Errorf("%s: 인증정보를 찾을 수 없습니다 - ev:\n%s", util.GetCurrentFuncName(), string(bytes))
		}

		ok := true
		// TODO: Lambda Authorizer 생성해야함
		//reqCtx.UserAuthRequestContext.UserID, ok = in.Event.RequestContext.Authorizer["claims"].(map[string]any)["cognito:username"].(string)
		if !ok {
			return nil, errors.Errorf("%s: 'user 정보를 찾을 수 없습니다", util.GetCurrentFuncName())
		}

		// for debug
		bytes, _ := json.MarshalIndent(&reqCtx, "", "    ")
		log.Printf("%s: reqCtx:\n%s\n", util.GetCurrentFuncName(), string(bytes))
		//

	case RequestAuthIAM:
		// IAM 인증
		reqCtx.AuthType = RequestAuthIAM
		reqCtx.IAMAuthRequestContext.UserArn = in.Event.RequestContext.Identity.UserArn
		reqCtx.IAMAuthRequestContext.User = in.Event.RequestContext.Identity.User
		reqCtx.IAMAuthRequestContext.AccessKey = in.Event.RequestContext.Identity.AccessKey
	default:
		bytes, _ := json.Marshal(in.Event)
		return nil, errors.Errorf("%s: 지원하지 않는 인증방식입니다 - ev:\n%s", util.GetCurrentFuncName(), string(bytes))
	}

	return &reqCtx, nil
}

//// InvokeGetAccessibleResourcesLambda : 가용자원목록을 반환하는 람다를 호출해서 결과를 반환해줌
//func InvokeGetAccessibleResourcesLambda(ctx context.Context, in *InvokeGetAccessibleResourcesLambdaInput) ([]string, error) {
//	if nil == in {
//		return nil, errors.New("입력파라미터가 nil입니다")
//	} else if err := in.Validate(); nil != err {
//		return nil, errors.Wrap(err, "입력파라미터가 잘못 입력되었습니다")
//	}
//
//	lambdaReq := GetAccessibleResourcesInput{
//		ServiceName: in.ServiceName,
//		UserID:      in.UserID,
//	}
//	reqBytes, err := json.Marshal(&lambdaReq)
//	if nil != err {
//		return nil, errors.Wrap(err, "람다 요청을 마살링하지 못했습니다")
//	}
//	out, err := in.LambdaClient.InvokeWithContext(ctx, &lambdaSvc.InvokeInput{
//		FunctionName:   aws.String(in.LambdaName),
//		InvocationType: aws.String("RequestResponse"),
//		Payload:        reqBytes,
//	})
//	if nil != err {
//		return nil, errors.Wrapf(err, "%s 람다를 호출하지 못했습니다", in.LambdaName)
//	} else if nil == out {
//		return nil, errors.Errorf("%s 람다 호출결과가 비어 있습니다", in.LambdaName)
//	}
//
//	var res GetAccessibleResourcesOutput
//	err = json.Unmarshal(out.Payload, &res)
//	if nil != err {
//		return nil, errors.Wrapf(err, "%s 람다 반환값을 복원하지 못했습니다", in.LambdaName)
//	}
//	return res.Resources, nil
//}
