package util

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/xid"
)

func init() {
	gob.Register(map[string]interface{}{})
}

// GetTypeName : 패키지명 제외하고 타입명만 추출
func GetTypeName(myvar interface{}) string {
	t := reflect.TypeOf(myvar)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	} else if t.Kind() == reflect.Slice {
		return "[]" + t.Elem().Name()
	}
	return t.Name()
}

// NewUUID :
func NewUUID() string {
	return xid.New().String()
}

// NewEntityID : 새 엔티티 아이디 생성
func NewEntityID(entityType string) string {
	return fmt.Sprintf("%s-%s", entityType, NewUUID())
}

// StringPointer ...
func StringPointer(v string) *string {
	return &v
}

// BoolPointer :
func BoolPointer(v bool) *bool {
	return &v
}

// TimePointer :
func TimePointer(v time.Time) *time.Time {
	return &v
}

// TimeToISOString
// format: 2023-01-27T19:16:07+09:00
func TimeToISOString(v time.Time) string {
	return v.Format(time.RFC3339Nano)
}

// TimeFromIsoString
func TimeFromIsoString(str string) (*time.Time, error) {
	t, err := time.Parse(time.RFC3339Nano, str)
	if nil != err {
		return nil, err
	}
	return &t, nil
}

// ToLowerCaseStringArray :
func ToLowerCaseStringArray(arr []string) {
	for i := range arr {
		arr[i] = strings.ToLower(arr[i])
	}
}

// SplitCSVWithoutEmptyString :
func SplitCSVWithoutEmptyString(csv string) []string {
	return SplitWithoutEmptyString(csv, ",")
}

// SplitWithoutEmptyString :
func SplitWithoutEmptyString(str, separator string) []string {
	rawArr := strings.Split(str, separator)
	arr := []string{}
	for _, v := range rawArr {
		if "" == strings.TrimSpace(v) {
			continue
		}
		arr = append(arr, v)
	}
	return arr
}

// ContainsInStringList : 문자열 배열에 문자열이 존재하는지 검사
func ContainsInStringList(strList []string, str string) bool {
	for _, v := range strList {
		if v == str {
			return true
		}
	}
	return false
}

// RemoveStringFromStringList : 문자열 배열에서 특정 문자열 항목을 제거하고 제거한 횟수를 리턴
func RemoveStringFromStringList(str string, strList *[]string) int {
	oldCount := len(*strList)
	i := 0
	for _, x := range *strList {
		if str != x {
			(*strList)[i] = x
			i++
		}
	}
	*strList = (*strList)[:i]
	return oldCount - i
}

// RemoveDuplicatesFromStringList : 문자열 배열에서 중복문자열을 제거하여 새 문자열 배열을 반환
func RemoveDuplicatesFromStringList(srcList []string) []string {
	keys := map[string]bool{}
	resultList := []string{}
	for _, v := range srcList {
		if _, exist := keys[v]; !exist {
			keys[v] = true
			resultList = append(resultList, v)
		}
	}
	return resultList
}

// IsExistDuplicatesInStringList : 문자열 배열에 중복이 있는지 검사
func IsExistDuplicatesInStringList(list []string) bool {
	keys := map[string]bool{}
	for _, v := range list {
		keys[v] = true
	}
	return len(list) != len(keys)
}

// EqualStringListIgnoreOrder : 문자열 배열을 순서 상관없이 비교
func EqualStringListIgnoreOrder(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	// create a map of string -> int
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		// 0 value for int is 0, so just increment a counter for the string
		diff[_x]++
	}
	for _, _y := range y {
		// If the string _y is not in diff bail out early
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y]--
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}

// func getCallerName() string {
// 	pc := make([]uintptr, 10) // at least 1 entry needed
// 	runtime.Callers(2, pc)
// 	f := runtime.FuncForPC(pc[0])
// 	names := strings.Split(f.Name(), ".")
// 	return names[len(names)-1]
// }

// GetCurrentFuncName :
func GetCurrentFuncName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	//log.Printf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
	return frame.Function
}

// CompressToBase64 : 바이트배열을 압축하여 base64 문자열로 반환
func CompressToBase64(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return "", err
	}
	if err := gz.Flush(); err != nil {
		return "", err
	}
	if err := gz.Close(); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

// UncompressFromBase64 : base64문자열을 디코딩한후 압축해제하여 반환
func UncompressFromBase64(base64Str string) ([]byte, error) {
	if "" == base64Str {
		return []byte{}, nil
	}
	decodedBuf, err := base64.StdEncoding.DecodeString(base64Str)
	if nil != err {
		return nil, err
	}
	rdata := bytes.NewReader(decodedBuf)
	r, err := gzip.NewReader(rdata)
	if nil != err {
		return nil, err
	}
	restoredData, err := ioutil.ReadAll(r)
	if nil != err {
		return nil, err
	}
	return restoredData, nil
}

// FindDuplicateItemsBetweenStringLists : 두 문자열 배열간에 중복되는 항목을 찾아 리스트로 반환
func FindDuplicateItemsBetweenStringLists(list1, list2 []string) []string {
	duplicates := []string{}
	for _, str := range list1 {
		if ContainsInStringList(list2, str) {
			duplicates = append(duplicates, str)
		}
	}
	return duplicates
}

// FindNotExistItemsInSrcStringList : 대상배열 중에 원본 배열에 없는 항목을 찾아 반환
func FindNotExistItemsInSrcStringList(src, dst []string) []string {
	notExistItems := []string{}
	for _, str := range dst {
		if !ContainsInStringList(src, str) {
			notExistItems = append(notExistItems, str)
		}
	}
	return notExistItems
}

// FindDifferenceItemsBetweenStringLists : 두 문자열 배열간에 중복되지 않는 항목을 찾아 리스트로 반환
func FindDifferenceItemsBetweenStringLists(list1, list2 []string) []string {
	diffItems := FindNotExistItemsInSrcStringList(list1, list2)
	diffItems = append(diffItems, FindNotExistItemsInSrcStringList(list2, list1)...)
	return diffItems
}

// CloneInterfaceMap : deep copy  - map[string]interface{}
func CloneInterfaceMap(m map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	err := enc.Encode(m)
	if err != nil {
		return nil, err
	}
	var copy map[string]interface{}
	err = dec.Decode(&copy)
	if err != nil {
		return nil, err
	}
	return copy, nil
}

// ObjectToMapInterface :
func ObjectToMapInterface(obj interface{}) (map[string]interface{}, error) {
	if nil == obj {
		return nil, errors.New("obj가 입력되지 않았습니다")
	}
	bytes, err := json.Marshal(obj)
	if nil != err {
		return nil, err
	}
	var mapIf map[string]interface{}
	err = json.Unmarshal(bytes, &mapIf)
	if nil != err {
		return nil, err
	}
	return mapIf, nil
}

// InterfaceMapToJSONRawMessage ...
func InterfaceMapToJSONRawMessage(ifMap map[string]interface{}) (json.RawMessage, error) {
	if len(ifMap) == 0 {
		return nil, nil
	}
	bytes, err := json.Marshal(ifMap)
	if nil != err {
		return nil, errors.Wrap(err, "Args입력파라미터를 마샬링하지 못했습니다")
	}
	return bytes, nil
}

// IsFileExists ...
func IsFileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
