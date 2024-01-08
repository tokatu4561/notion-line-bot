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

func(c *Client) GetPage(ctx context.Context, pageId string) (*notionapi.Page, error) {
	page, err := c.api.Page.Get(ctx, notionapi.PageID(pageId))
	if err != nil {
		return nil, err
	}

	return page, nil
}

func(c *Client) GetChildren(ctx context.Context, blockId string) ([]notionapi.Block, error) {
	children, err := c.api.Block.GetChildren(ctx, notionapi.BlockID(blockId), nil)
	if err != nil {
		return nil, err
	}

	return children.Results, nil
}