package main

import (
	"context"
	"fmt"
	"app/line"
	"app/notion"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
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

// 仮のlambda関数
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()

	line, err := line.SetUpClient()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "line connection error", StatusCode: 500}, err
	}

	lineEvents, err := line.ParseRequest(request)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "line parse error", StatusCode: 500}, err
	}

	db, err = setUpDB()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "db connection error", StatusCode: 500}, err
	}

	for _, event := range lineEvents {
		// イベントがメッセージの受信だった場合
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
				case *linebot.TextMessage:

					userId := event.Source.UserID

					// インテグレーションキーの登録
					if strings.Contains(message.Text, "キーを登録") {
						// 文字列からインテグレーションキーを抽出
						pattern := `integrate:(secret_[a-zA-Z0-9]+)`
						re := regexp.MustCompile(pattern)
						integrationKey := re.FindStringSubmatch(message.Text)[1]
						
						//databaseid を抽出
						pattern = `databaseId:([a-zA-Z0-9]+)`
						re = regexp.MustCompile(pattern)
						databaseId := re.FindStringSubmatch(message.Text)[1]

						// db に登録
						err = registrationNotionKey(userId, integrationKey, databaseId)
						if err != nil {
							return events.APIGatewayProxyResponse{
								Body:       err.Error(),
								StatusCode: 500,
							}, nil
						}
						
						// インテグレーションキーの登録メッセージを返信
						replyMessage := "インテグレーションキーを登録しました。"
						_, err = line.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do()

						if err != nil {
							return events.APIGatewayProxyResponse{
								Body:       err.Error(),
								StatusCode: 500,
							}, nil
						}
					} else {
						// ユーザーを特定
						var param LineNotionKeyParam
						err = db.Table(tableName).Get("line_id", userId).One(&param)
						if err != nil {
							return events.APIGatewayProxyResponse{
								Body:       err.Error(),
								StatusCode: 500,
							}, nil
						}

						// インテグレーションキーを取得
						integrationKey := param.IntegrationKey
						databaseId := param.DatabaseId

						// メッセージを Notion に送信
						notionClient, _ := notion.NewClient(integrationKey)
						notionClient.AppendText(ctx, databaseId, message.Text)

						// line メッセージを返信
						replyMessage := fmt.Sprintf("メッセージ「%s」を Notion に送信しました。", message.Text)
						_, err = line.Client.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do() 

						if err != nil {
							return events.APIGatewayProxyResponse{
								Body:       err.Error(),
								StatusCode: 500,
							}, nil
						}
					}
				default:
			}
		}
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello, %v", string("hello")),
		StatusCode: 200,
	}, nil
}

// registration integration key and pageId to db 
func registrationNotionKey(userId string, integrationKey string, blockId string) error {
	param := LineNotionKeyParam{
		LineId: userId,
		IntegrationKey: integrationKey,
		DatabaseId: blockId,
	}

	err := db.Table(tableName).Put(param).Run()
	if err != nil {
		return err
	}

	return nil
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