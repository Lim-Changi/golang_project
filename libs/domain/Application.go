package domain

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Application :
type Application interface {
	ExecutableName() string
	Revision() string
	Version() string
	BuildTime() time.Time // utc
	Print()
}

// application :
type application struct {
	revision  string
	version   string
	buildTime time.Time
}

func (t application) ExecutableName() string {
	return filepath.Base(os.Args[0])
}

func (t application) Revision() string {
	return t.revision
}
func (t application) Version() string {
	return t.version
}
func (t application) BuildTime() time.Time {
	return t.buildTime
}

func (t application) Print() {
	log.Printf("ExecutalbeName: %s\n", t.ExecutableName())
	log.Printf("Version: %s\n", t.version)
	log.Printf("Revision: %s\n", t.revision)
	log.Printf("BuildTime: %s\n", t.buildTime.Format("2006-01-02T15:04:05Z"))
}

// NewApplication :
// "2006-01-02T15:04:05Z", date -u +%Y-%m-%dT%H:%M:%SZ
// buildTimeISO8601 : ex)"2020-03-03T03:02:05Z"
func NewApplication(revision, version, buildTimeISO8601 string) (Application, error) {
	layout := "2006-01-02T15:04:05Z"
	buildTime, err := time.Parse(layout, buildTimeISO8601)
	if err != nil {
		return nil, err
	}
	return &application{
		revision:  revision,
		version:   version,
		buildTime: buildTime,
	}, nil
}

// MethodInfo ...
type MethodInfo struct {
	Description       string   `jsonschema:"description=API 함수설명"`
	RequestSchema     string   `jsonschema:"description=요청파라미터 JSON 스키마"`
	ResponseSchema    string   `jsonschema:"description=응답파라미터 JSON 스키마"`
	AllowedRequesters []string `jsonschema:"description=허용된 요청자 타입목록"`
}
