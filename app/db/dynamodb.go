package db

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

const AWS_REGION = "ap-northeast-1"
const tableName = "line-notion-keys"

type Repository struct {
	db *dynamo.DB
}

func NewDynamoDatabaseHandler() *dynamo.DB {
	sess, _ := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
	})

	db := dynamo.New(sess)

	return db
}

func NewRepository() RepositoryInterface {
	return &Repository{
		db: NewDynamoDatabaseHandler(),
	}
}

func (r *Repository) Add(line_id string, integration_key string, database_id string) error {
	param := LineNotionKey{
		LineId: line_id,
		IntegrationKey: integration_key,
		DatabaseId: database_id,
	}

	err := r.db.Table(tableName).Put(param).Run()
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) Get(line_id string) (*LineNotionKey, error) {
	var param LineNotionKey

	err := r.db.Table(tableName).Get("line_id", line_id).One(&param)
	if err != nil {
		return nil, err
	}

	return &param, nil
}

func (r *Repository) GetList() ([]*LineNotionKey, error) {
	var params []*LineNotionKey

	err := r.db.Table(tableName).Scan().All(&params)
	if err != nil {
		return nil, err
	}

	return params, nil
}