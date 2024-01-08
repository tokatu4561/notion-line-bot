package db

type LineNotionKey struct {
	LineId string `dynamo:"line_id" json:"line_id"`
	IntegrationKey string `dynamo:"integration_key" json:"integration_key"`
	DatabaseId string `dynamo:"database_id" json:"database_id"`
}

type RepositoryInterface interface {
	Add(line_id string, integration_key string, database_id string) error
	Get(line_id string) (*LineNotionKey, error)
	GetList() ([]*LineNotionKey, error)
}