package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/sirupsen/logrus"
	"awesomeProject6/internal/models"
	"time"
)

type Client struct {
	es     *elasticsearch.Client
	index  string
	logger *logrus.Logger
}

func NewClient(urls []string, username, password, index string) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: urls,
		Username:  username,
		Password:  password,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()

	client := &Client{
		es:     es,
		index:  index,
		logger: logger,
	}

	if err := client.createIndexIfNotExists(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *Client) createIndexIfNotExists() error {
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"timestamp": map[string]interface{}{
					"type": "date",
				},
				"level": map[string]interface{}{
					"type": "keyword",
				},
				"message": map[string]interface{}{
					"type": "text",
				},
				"service": map[string]interface{}{
					"type": "keyword",
				},
				"host": map[string]interface{}{
					"type": "keyword",
				},
				"tags": map[string]interface{}{
					"type": "object",
				},
				"fields": map[string]interface{}{
					"type": "object",
				},
			},
		},
	}

	body, err := json.Marshal(mapping)
	if err != nil {
		return err
	}

	req := esapi.IndicesCreateRequest{
		Index: c.index,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(context.Background(), c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 400 {
		return fmt.Errorf("failed to create index: %s", res.Status())
	}

	return nil
}

func (c *Client) IndexLog(ctx context.Context, log models.LogEntry) error {
	body, err := json.Marshal(log)
	if err != nil {
		return err
	}

	indexName := fmt.Sprintf("%s-%s", c.index, time.Now().Format("2006.01.02"))

	req := esapi.IndexRequest{
		Index: indexName,
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index log: %s", res.Status())
	}

	return nil
}

func (c *Client) BulkIndexLogs(ctx context.Context, logs []models.LogEntry) error {
	var buf bytes.Buffer

	for _, log := range logs {
		indexName := fmt.Sprintf("%s-%s", c.index, time.Now().Format("2006.01.02"))
		
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": indexName,
			},
		}
		metaBytes, _ := json.Marshal(meta)
		buf.Write(metaBytes)
		buf.WriteByte('\n')

		logBytes, err := json.Marshal(log)
		if err != nil {
			return err
		}
		buf.Write(logBytes)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body: &buf,
	}

	res, err := req.Do(ctx, c.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index failed: %s", res.Status())
	}

	return nil
}