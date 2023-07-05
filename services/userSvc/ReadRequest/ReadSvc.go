package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-xray-sdk-go/xray"
	"log"
	"net/http"
)

func init() {
	log.SetFlags(0)
	// 람다 핸들러에서는 자동으로 세그먼트가 만들어져 컨텍스트로 전달되지만
	// 람다 초기화시 사용하려면 수동으로 세그먼트를 만들어줘야 한다.
	ctx, seg := xray.BeginSegment(context.Background(), "init")
	_ = ctx
	defer seg.Close(nil)

	awsSession := session.Must(session.NewSession())
	awsSession = xray.AWSSession(awsSession)

}

// Response :
type Response events.APIGatewayProxyResponse

// 서비스 메서드 호출시 HTTP 경로에서 인증타입과 요청명 추출
func lambdaHandler(ctx context.Context, ev *events.APIGatewayProxyRequest) (Response, error) {
	if nil == ev {
		return Response{}, nil
	}
	if "" == ev.Path {
		return Response{}, nil
	}
	log.Printf(">> Path:\n %s\n", ev.Path) // {+proxy}부분에 들어오는 문자열 출력, ex) /read/GetCatalogListRequest
	log.Printf(">> Body:\n %s\n", ev.Body) // 본문 JSON 문자열
	log.Printf(">> event:\n %v\n", *ev)

	//// readRequest 조립터
	//// Path = "[svc]/[admin/user/sys]/read/XXXXXXRequest"
	//httpRequestPath, err := httppath.ParseHTTPRequestPath(ev.Path)
	//if nil != err {
	//	err = errors.Wrap(err, "요청 경로 파싱에 실패하였습니다")
	//	return Response{StatusCode: http.StatusBadRequest, Body: err.Error()}, nil
	//} else if nil == httpRequestPath {
	//	err = errors.New("요청 경로 파싱 결과 빈 결과를 반환했습니다")
	//	return Response{StatusCode: http.StatusBadRequest, Body: err.Error()}, nil
	//} else if requests.RequestMethodRead != httpRequestPath.RequestMethodType {
	//	err = errors.Errorf("요청 경로 중 요청 타입이 잘못 입력되었습니다 - expected: %s, actual: %s", requests.RequestMethodRead, httpRequestPath.RequestMethodType)
	//	return Response{StatusCode: http.StatusBadRequest, Body: err.Error()}, nil
	//}

	//var readRequest *commonRead.RawRequest = &commonRead.RawRequest{
	//	Name:  httpRequestPath.RequestName,
	//	Param: []byte(ev.Body),
	//}
	//
	//// RequestContext 추출
	//reqCtx, err := domain.RetrieveAuthorizerRequestContext(ctx, &domain.RetrieveAuthorizerRequestContextInput{
	//	AuthType:      httpRequestPath.AuthType,
	//	RequesterType: httpRequestPath.RequesterType,
	//	//RequesterType: auth.RequesterSystem, // For Debug Purpose
	//	Event:       ev,
	//	ServiceName: cfg.SVCName,
	//	GetAccessibleResourcePRNListFunc: func(ctx context.Context, userID, serviceName string) ([]string, error) {
	//		return nil, nil
	//	},
	//})
	//if nil != err {
	//	err = errors.Wrap(err, "ReadSvc: RequestContext를 추출하지 못했습니다")
	//	return Response{
	//		StatusCode:      http.StatusInternalServerError,
	//		IsBase64Encoded: false,
	//		Body:            err.Error(),
	//	}, nil
	//}
	//
	//// ?? 컨텍스트에 심을 것인가? 파라미터로 전달할 것인가?
	//// 하위 루틴에서 로그 등을 남기는 용도로 사용할때 컨텍스트가 좋을 것!!!!
	//res, err := handler.Process(domain.EmbedRequestContext(ctx, reqCtx), readRequest)
	//if nil != err {
	//	log.Printf("%s: error- %s\n", util.GetCurrentFuncName(), err.Error())
	//	return Response{
	//		StatusCode:      http.StatusInternalServerError,
	//		IsBase64Encoded: false,
	//		Body:            err.Error(),
	//	}, nil //err
	//}
	//resBytes, err := json.Marshal(res)
	//fmt.Println("Response Body:::" + string(resBytes))
	//if nil != err {
	//	log.Printf("%s: error- %s\n", util.GetCurrentFuncName(), err.Error())
	//	return Response{
	//		StatusCode:      http.StatusInternalServerError,
	//		IsBase64Encoded: false,
	//		Body:            err.Error(),
	//	}, nil
	//}
	// json.HTMLEscape()
	return Response{
		StatusCode:      http.StatusOK,
		IsBase64Encoded: false,
		Body:            "test",
	}, nil
}

// LambdaHandler : for test purpose
func LambdaHandler(ctx context.Context, ev *events.APIGatewayProxyRequest) (Response, error) {
	return lambdaHandler(ctx, ev)
}

func main() {
	lambda.Start(lambdaHandler)
}
