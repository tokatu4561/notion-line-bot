package main

import (
	"context"
	"fmt"
	"lambda/line"
	"lambda/notion"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"github.com/line/line-bot-sdk-go/linebot"
)

type MyEvent struct {
	Name string `json:"name"`
}

const AWS_REGION = "ap-northeast-1"
const DYNAMO_ENDPOINT = "http://dynamodb:8000"
const tableName = "line-notion-keys"

var db *dynamo.DB

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
						pattern := `integrate:\s+([\w\d]+)`
						re := regexp.MustCompile(pattern)
						integrationKey := re.FindStringSubmatch(message.Text)[1]

						//pageid を抽出
						pattern = `databaseId:\s+([\w\d]+)`
						re = regexp.MustCompile(pattern)
						pageId := re.FindStringSubmatch(message.Text)[1]

						// db に登録
						registrationNotionKey(userId, integrationKey, pageId)
						
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
						var param interface{}
						err = db.Table(tableName).Get("user_id", userId).One(&param)
						if err != nil {
							return events.APIGatewayProxyResponse{
								Body:       err.Error(),
								StatusCode: 500,
							}, nil
						}

						// インテグレーションキーを取得
						integrationKey := param.(map[string]interface{})["integration_key"].(string)
						pageId := param.(map[string]interface{})["page_id"].(string)

						// メッセージを Notion に送信
						notionClient, _ := notion.NewClient(integrationKey)
						notionClient.AppendText(ctx, pageId, message.Text)

						// line メッセージを返信
						replyMessage := message.Text
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

type LineNotionKeyRegistrationParam struct {
	UserId string `dynamo:"user_id" json:"user_id"`
	IntegrationKey string `dynamo:"integration_key" json:"integration_key"`
	PageId string `dynamo:"page_id" json:"page_id"`
}

// registration integration key and pageId to db 
func registrationNotionKey(userId string, integrationKey string, pageId string) error {
	param := LineNotionKeyRegistrationParam{
		UserId: userId,
		IntegrationKey: integrationKey,
		PageId: pageId,
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
		Endpoint:    aws.String(DYNAMO_ENDPOINT),
		Credentials: credentials.NewStaticCredentials("dummy", "dummy", "dummy"),
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