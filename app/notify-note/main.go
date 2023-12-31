package main

import (
	"app/db"
	"app/line"
	"app/notion"
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/jomei/notionapi"
	"github.com/line/line-bot-sdk-go/linebot"
)

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	ctx := context.Background()

	line, err := line.SetUpClient()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "line connection error", StatusCode: 500}, err
	}

	db := db.NewRepository()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "db connection error", StatusCode: 500}, err
	}

	// db から全ユーザーのintegration key、line の情報を取得
	keys, err := db.GetList()
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "db scan error", StatusCode: 500}, err
	}

	// 各ユーザーの notion 上のメモを取得し、line に通知
	for _, key := range keys {
		// notion からメモを取得
		notionClient, _ := notion.NewClient(key.IntegrationKey)
		page, err := notionClient.GetPage(ctx, key.DatabaseId)
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
			// メモの内容が空の場合はスキップ
			if len(paragpagh.Paragraph.RichText) == 0 {
				continue
			}
			text := paragpagh.Paragraph.RichText[0].Text.Content
			messages = append(messages, text)
		}

		// line に通知
		textMessage := "メモリストの内容\n"
		for _, message := range messages {
			textMessage += message + "\n"
		}
		_, err = line.Client.PushMessage(key.LineId, linebot.NewTextMessage(textMessage)).Do()
		if err != nil {
			return events.APIGatewayProxyResponse{Body: "line push message error", StatusCode: 500}, err
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