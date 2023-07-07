package ddb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// EntityRepository : 엔티티 저장소
type EntityRepository struct {
	tableName     string
	db            *dynamodb.DynamoDB
	gsiIndexName  string
	lsiIndexName  string
	recordTTLDays int
}

// NewEntityRepository :
func NewEntityRepository(db *dynamodb.DynamoDB, tableName, gsiIndexName string, lsiIndexName string, recordTTLDays int) (*EntityRepository, error) {
	return &EntityRepository{
		db:            db,
		tableName:     tableName,
		gsiIndexName:  gsiIndexName,
		lsiIndexName:  lsiIndexName,
		recordTTLDays: recordTTLDays,
	}, nil
}
