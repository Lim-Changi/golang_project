package ddb

import (
	"context"
	"encoding/json"
	"fmt"
	"golang_project/libs/infra/util"
	"log"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/pkg/errors"
)

// DDB constants
const (
	BatchGetItemCountLimit       = 100
	BatchWriteItemCountLimit     = 25
	TransactWriteItemsCountLimit = 100
)

// NewDDBConnection :
func NewDDBConnection(awsSession *session.Session) *dynamodb.DynamoDB {
	// for debug
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf(">> NewDBConnection elapsed: %s \n", elapsed)
	}()
	c := dynamodb.New(awsSession)
	return c
}

// NewDDBConnectionWithRegion :
func NewDDBConnectionWithRegion(awsSession *session.Session, region string) *dynamodb.DynamoDB {
	// for debug
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Printf(">> NewDBConnection elapsed: %s \n", elapsed)
	}()
	c := dynamodb.New(awsSession, aws.NewConfig().WithRegion(region))
	return c
}

// DBConfig :
type DBConfig struct {
	TableName                                 string `jsonschema:"description=테이블명"`
	GSIIndexName                              string `jsonschema:"description=GSI 인덱스명,example:SK-DATA-index"`
	LSIIndexName                              string `jsonschema:"description=LSI 인덱스명,example:LSK-index"`
	RecordTTLDays                             int    `jsonschema:"description=레코드 TTL 수명일"`
	RecordTTLDaysForReportServiceVersionTopic int    `jsonschema:"description=ReportServiceVersionTopic 명령레코드 TTL 수명일"`
}

// LoadDBConfig : "TableName", "GSIIndexName", "LSIIndexName", "RecordTTLDays", "RecordTTLDaysForReportServiceVersionTopic"
func LoadDBConfig() *DBConfig {
	recordTTLDays, err := strconv.Atoi(os.Getenv("RecordTTLDays"))
	if nil != err {
		panic(errors.Wrap(err, "Failed to load DBConfig - RecordTTLDays"))
	}

	strNum, isExist := os.LookupEnv("RecordTTLDaysForReportServiceVersionTopic")
	var recordTTLDaysForReportServiceVersionTopic int = 0
	if isExist {
		recordTTLDaysForReportServiceVersionTopic, err = strconv.Atoi(strNum)
		_ = recordTTLDaysForReportServiceVersionTopic
		_ = err
		//log.Println("Cannot find to load DBConfig - RecordTTLDaysForReportServiceVersionTopic")
	} else {
		recordTTLDaysForReportServiceVersionTopic = recordTTLDays
	}

	return &DBConfig{
		TableName:     os.Getenv("TableName"),
		GSIIndexName:  os.Getenv("GSIIndexName"),
		LSIIndexName:  os.Getenv("LSIIndexName"),
		RecordTTLDays: recordTTLDays,
		RecordTTLDaysForReportServiceVersionTopic: recordTTLDaysForReportServiceVersionTopic,
	}
}

// BatchTransactWriteItems : TransactWriteItems의 최대 허용 수를 넘을 경우 최대 허용수로 분할하여 동시에 실행한다.
// 실패할 경우 성공한 다른 배치 작업에 대한 롤백은 없음.
// 참조: TransactWriteItems와 BatchWriteItem는 최대 25개? 까지만 가능하다.(25개 이상이 되려면 서비스 수가 24개 이상이 되어야 한다.)
// 프리셋 엔티티 레코드 업데이트와 함수세트 레코드들의 업데이트가 25개 이하이면 TransactWriteItems 한번으로 처리하고
// 25개 이상이면 TransactWriteItems을 여러번 수행한다.(BatchWriteItem은 PUT과 DELETE만 지원,TransactWriteItems는 PUT, UPDATE,DELETE )
func BatchTransactWriteItems(ctx context.Context, db *dynamodb.DynamoDB, writeItems []*dynamodb.TransactWriteItem) error {
	// TransactWriteItems를 몇번 수행할지 계산
	batchCount := int(math.Ceil(
		float64(len(writeItems)) / float64(TransactWriteItemsCountLimit)))

	wg := sync.WaitGroup{}
	errorList := make([]error, batchCount)
	for i := 0; i < batchCount; i++ {
		startIdx := i * TransactWriteItemsCountLimit
		var endIdx int
		if i < batchCount-1 {
			// 마지막 배치가 아니면 배치사이즈 만큼의 인덱스
			endIdx = (i + 1) * TransactWriteItemsCountLimit
		} else {
			// 마지막 배치이면 원본배열의 마지막 인덱스까지
			endIdx = len(writeItems)
		}
		input := &dynamodb.TransactWriteItemsInput{
			TransactItems: writeItems[startIdx:endIdx],
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errorList[i] = db.TransactWriteItemsWithContext(ctx, input)
		}(i)
	}
	wg.Wait()
	for i, e := range errorList {
		if nil != e {
			return errors.Wrapf(e, "BatchTransactWriteItems- %d번째 배치작업이 실패하였습니다", i+1)
		}
	}
	return nil
}

// BatchWriteItemsForOneTable : 특정 하나의 테이블에 대한 배치 쓰기 작업
// 처리되지 못한 작업목록과 에러 목록 반환
func BatchWriteItemsForOneTable(ctx context.Context, db *dynamodb.DynamoDB, tableName string, writeRequests []*dynamodb.WriteRequest) (unprocessedItems []*dynamodb.WriteRequest, err error) {
	// BatchWriteItems를 몇번 수행할지 계산
	batchCount := int(math.Ceil(
		float64(len(writeRequests)) / float64(BatchWriteItemCountLimit)))

	wg := sync.WaitGroup{}
	errList := make([]error, batchCount)
	outList := make([]*dynamodb.BatchWriteItemOutput, batchCount)
	writeRequestsList := make([][]*dynamodb.WriteRequest, batchCount)
	for i := 0; i < batchCount; i++ {
		startIdx := i * BatchWriteItemCountLimit
		var endIdx int
		if i < batchCount-1 {
			// 마지막 배치가 아니면 배치사이즈 만큼의 인덱스
			endIdx = (i + 1) * BatchWriteItemCountLimit
		} else {
			// 마지막 배치이면 원본배열의 마지막 인덱스까지
			endIdx = len(writeRequests)
		}
		writeRequestsList[i] = writeRequests[startIdx:endIdx]

		input := &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]*dynamodb.WriteRequest{
				tableName: writeRequestsList[i],
			},
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outList[i], errList[i] = db.BatchWriteItemWithContext(ctx, input)
		}(i)
	}
	wg.Wait()

	for i := 0; i < batchCount; i++ {
		if nil != errList[i] {
			unprocessedItems = append(unprocessedItems, writeRequestsList[i]...)
			continue
		}
		if nil == outList[i] {
			continue
		}
		if len(outList[i].UnprocessedItems) == 0 {
			continue
		}
		unprocessedItems = append(unprocessedItems, outList[i].UnprocessedItems[tableName]...)
	}
	errStr := ""
	for i, e := range errList {
		if nil != e {
			errStr += errors.Wrapf(e, "%s: %d번째 배치작업이 실패하였습니다\n", util.GetCurrentFuncName(), i+1).Error()
			log.Printf("%s: 실패한 작업- %d번째 : %v\n", util.GetCurrentFuncName(), i+1, *writeRequests[i])
		}
	}
	if errStr != "" {
		err = errors.New(errStr)
	}
	return unprocessedItems, err
}

// MakeStringSliceToINStatementExpressionAndValuesInput ...
type MakeStringSliceToINStatementExpressionAndValuesInput struct {

	// KeyName : IN 구문의 키이름 예) "ResourceType"
	KeyName string
	// ExpressionKeyPrefix: IN 구문 괄호안에 들어갈 키워드
	// 예) 키워드가 "resType"일 경우 => ResourceType IN ( :resType1, :resType2, :resType3, ...)
	ExpressionKeyPrefix string
	// IN 구문 괄호안에 들어가 실제 문자열 배열값
	// 예) StringValues: {"a", "b", "c"}
	StringValues []string
}

// Validate ...
func (t MakeStringSliceToINStatementExpressionAndValuesInput) Validate() error {
	if t.KeyName == "" {
		return errors.New("KeyName이 비어있습니다")
	}
	if t.ExpressionKeyPrefix == "" {
		return errors.New("ExpressionKeyPrefix가 비어있습니다")
	}
	if len(t.StringValues) == 0 {
		return errors.New("StringValues가 비어있습니다")
	}
	return nil
}

// MakeStringSliceToINStatementExpressionAndValuesOutput ...
type MakeStringSliceToINStatementExpressionAndValuesOutput struct {
	// Expression: FilterExpression 따위에 사용할 완성된 In 구문 문자열
	// 예) "ResourceType IN (:resType1, :resType2, :resType3, ...)"
	Expression string
	// IN 구문 괄호안에 들어갈 실제 속성값 맵
	// Values := map[string]*dynamodb.AttributeValue{
	// 	":resType1":{
	// 		S: aws.String("a"),
	// 	},
	// 	":resType2":{
	// 		S: aws.String("b"),
	// 	},
	// 	":resType3":{
	// 		S: aws.String("c"),
	// 	},
	// }
	Values map[string]*dynamodb.AttributeValue
}

// MakeStringSliceToINStatementExpressionAndValues : FilterExpression 따위에 사용되는 IN 구문 문자열과 속성맵을 만들어 반환
func MakeStringSliceToINStatementExpressionAndValues(in *MakeStringSliceToINStatementExpressionAndValuesInput) (*MakeStringSliceToINStatementExpressionAndValuesOutput, error) {
	if nil == in {
		return nil, errors.Errorf("%s: 입력값이 nil입니다", util.GetCurrentFuncName())
	} else if err := in.Validate(); nil != err {
		return nil, errors.Wrapf(err, "%s: 입력값 검사가 실패하였습니다", util.GetCurrentFuncName())
	}
	out := &MakeStringSliceToINStatementExpressionAndValuesOutput{
		Values: map[string]*dynamodb.AttributeValue{},
	}
	var keyCSV string
	for i, stringValue := range in.StringValues {
		key := fmt.Sprintf(":%s%d", in.ExpressionKeyPrefix, i+1)
		if i > 0 {
			keyCSV += ","
		}
		keyCSV += key
		out.Values[key] = &dynamodb.AttributeValue{
			S: aws.String(stringValue),
		}
	}
	// 예) ResourceType IN (:resType1, :resType2, :resType3, ...)
	out.Expression = fmt.Sprintf("%s IN (%s)", in.KeyName, keyCSV)
	return out, nil
}

// MakePutTransactWriteRequest ...
func MakePutTransactWriteRequest(tableName string, recordPtr interface{}) (*dynamodb.TransactWriteItem, error) {
	if nil == recordPtr {
		return nil, errors.New("item이 nil 입니다")
	}
	av, err := dynamodbattribute.MarshalMap(recordPtr)
	if nil != err {
		return nil, err
	}
	return &dynamodb.TransactWriteItem{
		Put: &dynamodb.Put{
			TableName: aws.String(tableName),
			Item:      av,
		},
	}, nil
}

// MakePutBatchWriteRequest :
func MakePutBatchWriteRequest(recordPtr interface{}) (*dynamodb.WriteRequest, error) {
	if nil == recordPtr {
		return nil, errors.New("item이 nil 입니다")
	}
	av, err := dynamodbattribute.MarshalMap(recordPtr)
	if nil != err {
		return nil, err
	}
	return &dynamodb.WriteRequest{
		PutRequest: &dynamodb.PutRequest{
			Item: av,
		},
	}, nil
}

// MakePutBatchWriteRequests :
func MakePutBatchWriteRequests(recordPtrList []interface{}) ([]*dynamodb.WriteRequest, error) {
	if len(recordPtrList) == 0 {
		return nil, nil
	}
	results := []*dynamodb.WriteRequest{}
	for i := range recordPtrList {
		item, err := MakePutBatchWriteRequest(recordPtrList[i])
		if nil != err {
			return nil, err
		}
		results = append(results, item)
	}
	return results, nil
}

// BatchGetItemsForOneTableInput ..
type BatchGetItemsForOneTableInput struct {
	DB        *dynamodb.DynamoDB
	TableName string
	Keys      []map[string]*dynamodb.AttributeValue
}

// Validate ...
func (t BatchGetItemsForOneTableInput) Validate() error {
	if nil == t.DB {
		return errors.New("DB가 nil입니다")
	}
	if t.TableName == "" {
		return errors.New("TableName이 입력되지 않았습니다")
	}
	if len(t.Keys) == 0 {
		return errors.New("Keys가 입력되지 않았습니다")
	}
	return nil
}

// BatchGetItemsForOneTableOutput ...
type BatchGetItemsForOneTableOutput struct {
	Items           []map[string]*dynamodb.AttributeValue
	UnprocessedKeys []map[string]*dynamodb.AttributeValue
}

// BatchGetItemsForOneTable : 특정 하나의 테이블에 대한 배치 조회 작업, BatchGetItems 제한수와 관계없이 병렬작업으로 처리
func BatchGetItemsForOneTable(ctx context.Context, in *BatchGetItemsForOneTableInput) (*BatchGetItemsForOneTableOutput, error) {
	if nil == in {
		return nil, errors.New("입력 파라미터가 nil입니다")
	} else if err := in.Validate(); nil != err {
		return nil, errors.Wrapf(err, "%s: 입력파라미터 검사가 실패하였습니다", util.GetCurrentFuncName())
	}
	// BatchGetItems를 몇번 수행할지 계산
	batchCount := int(math.Ceil(
		float64(len(in.Keys)) / float64(BatchGetItemCountLimit)))

	wg := sync.WaitGroup{}
	errList := make([]error, batchCount)
	outList := make([]*dynamodb.BatchGetItemOutput, batchCount)
	keysList := make([][]map[string]*dynamodb.AttributeValue, batchCount)
	for i := 0; i < batchCount; i++ {
		startIdx := i * BatchGetItemCountLimit
		var endIdx int
		if i < batchCount-1 {
			// 마지막 배치가 아니면 배치사이즈 만큼의 인덱스
			endIdx = (i + 1) * BatchGetItemCountLimit
		} else {
			// 마지막 배치이면 원본배열의 마지막 인덱스까지
			endIdx = len(in.Keys)
		}
		keysList[i] = in.Keys[startIdx:endIdx]

		input := &dynamodb.BatchGetItemInput{
			RequestItems: map[string]*dynamodb.KeysAndAttributes{
				in.TableName: {
					Keys: keysList[i],
				},
			},
		}
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			outList[i], errList[i] = in.DB.BatchGetItemWithContext(ctx, input)
		}(i)
	}
	wg.Wait()

	outItems := []map[string]*dynamodb.AttributeValue{}
	unprocessedKeys := []map[string]*dynamodb.AttributeValue{}
	for i := 0; i < batchCount; i++ {
		// 에러가 하나라도 있으면 바로 리턴
		if err := errList[i]; nil != err {
			keyListBytes, _ := json.Marshal(keysList[i])
			return nil, errors.Wrapf(
				err,
				"%s: %d번째 배치작업이 실패하였습니다 - %d번째 키 목록 : %s\n",
				util.GetCurrentFuncName(),
				i+1,
				string(keyListBytes))
		}
		if nil == outList[i] {
			keyListBytes, _ := json.Marshal(keysList[i])
			return nil, errors.Errorf(
				"%s: %d번째 배치작업이 nil을 반환하였습니다 - %d번째 키 목록 : %s\n",
				util.GetCurrentFuncName(),
				i+1,
				string(keyListBytes))
		}
		if nil != outList[i].UnprocessedKeys[in.TableName] && len(outList[i].UnprocessedKeys[in.TableName].Keys) > 0 {
			unprocessedKeys = append(unprocessedKeys, outList[i].UnprocessedKeys[in.TableName].Keys...)
		}
		outItems = append(outItems, outList[i].Responses[in.TableName]...)
	}

	return &BatchGetItemsForOneTableOutput{
		Items:           outItems,
		UnprocessedKeys: unprocessedKeys,
	}, nil
}
