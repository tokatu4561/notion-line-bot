package main

import (
	"app/db"
	"app/line"
	"app/notion"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/line/line-bot-sdk-go/linebot"
)

type MyEvent struct {
	Name string `json:"name"`
}

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

	db := db.NewRepository()
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
						db.Add(userId, integrationKey, databaseId)
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
						user, err := db.Get(userId)
						if err != nil {
							return events.APIGatewayProxyResponse{
								Body:       err.Error(),
								StatusCode: 500,
							}, nil
						}

						// インテグレーションキーを取得
						integrationKey := user.IntegrationKey
						databaseId := user.DatabaseId

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

func main() {
	lambda.Start(handler)
}