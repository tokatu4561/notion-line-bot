package notion

import (
	"context"
	"errors"

	"github.com/jomei/notionapi"
)

type Client struct {
	api *notionapi.Client
}

// NewClient returns a new Notion API client.
func NewClient(token string) (*Client, error) {
	if token == "" {
		return nil, errors.New("notion API token is empty")
	}

	notionToken := notionapi.Token(token)

	return &Client{
		api: notionapi.NewClient(notionToken),
	}, nil
}

func(c *Client) AppendText(ctx context.Context, blockId string, text string) error {
	appendBlockChildrenParam := notionapi.AppendBlockChildrenRequest{
		Children: []notionapi.Block{
			&notionapi.ParagraphBlock{
				BasicBlock: notionapi.BasicBlock{
					Object: notionapi.ObjectTypeBlock,
					Type:   notionapi.BlockTypeParagraph,
				},
				Paragraph: notionapi.Paragraph{
					RichText: []notionapi.RichText{
						{
							Type: notionapi.ObjectTypeText,
							Text: &notionapi.Text{
								Content: text,
							},
						},
					},
				},
			},
		},
	}

	_, err := c.api.Block.AppendChildren(ctx, notionapi.BlockID(blockId), &appendBlockChildrenParam)
	if err != nil {
		return err
	}

	return nil
}