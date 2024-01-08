package main

import (
	"app/line"
	"app/notion"
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/jomei/notionapi"
	"github.com/line/line-bot-sdk-go/linebot"
)

type MyEvent struct {
	Name string `json:"name"`
}

const AWS_REGION = "ap-northeast-1"
const tableName = "line-notion-keys"

var db *dynamo.DB

type LineNotionKeyParam struct {
	LineId string `dynamo:"line_id" json:"line_id"`
	IntegrationKey string `dynamo:"integration_key" json:"integration_key"`
	DatabaseId string `dynamo:"database_id" json:"database_id"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()

	line, err := line.SetUpClient()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "line connection error", StatusCode: 500}, err
	}

	db, err = setUpDB()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "db connection error", StatusCode: 500}, err
	}

	// db から全ユーザーのintegration key、line の情報を取得
	var params []LineNotionKeyParam
	err = db.Table(tableName).Scan().All(&params)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "db scan error", StatusCode: 500}, err
	}

	// 各ユーザーの notion 上のメモを取得し、line に通知
	for _, param := range params {
		// notion からメモを取得
		notionClient, _ := notion.NewClient(param.IntegrationKey)
		page, err := notionClient.GetPage(ctx, param.DatabaseId)
		if err != nil {
			return events.APIGatewayProxyResponse{Body: "notion get page error", StatusCode: 500}, err
		}
		blocks, err := notionClient.GetChildren(ctx, string(page.ID))
		if err != nil {
			return events.APIGatewayProxyResponse{Body: "notion get children error", StatusCode: 500}, err
		}
		var messages []string
		for _, block := range blocks {
			paragpagh := block.(*notionapi.ParagraphBlock)
			text := paragpagh.Paragraph.RichText[0].Text.Content
			messages = append(messages, text)
		}

		// line に通知
		textMessage := ""
		for _, message := range messages {
			textMessage += message + "\n"
		}
		_, err = line.Client.PushMessage(param.LineId, linebot.NewTextMessage(textMessage)).Do()
		if err != nil {
			return events.APIGatewayProxyResponse{Body: "line push message error", StatusCode: 500}, err
		}
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", string("hello")),
		StatusCode: 200,
	}, nil
}

func setUpDB() (*dynamo.DB, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(AWS_REGION),
	})
	if err != nil {
		return nil, err
	}

	db := dynamo.New(sess)

	return db, nil
}

func main() {
	lambda.Start(handler)
}