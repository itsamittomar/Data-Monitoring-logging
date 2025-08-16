package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"awesomeProject6/internal/models"
)

type Consumer struct {
	consumer sarama.ConsumerGroup
	topics   []string
	handler  *ConsumerGroupHandler
	logger   *logrus.Logger
}

type ConsumerGroupHandler struct {
	logChan chan models.LogEntry
	logger  *logrus.Logger
}

func NewConsumer(brokers []string, groupID string, topics []string, logChan chan models.LogEntry) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	logger := logrus.New()
	handler := &ConsumerGroupHandler{
		logChan: logChan,
		logger:  logger,
	}

	return &Consumer{
		consumer: consumer,
		topics:   topics,
		handler:  handler,
		logger:   logger,
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err := c.consumer.Consume(ctx, c.topics, c.handler)
			if err != nil {
				c.logger.Errorf("Error consuming messages: %v", err)
				return err
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.consumer.Close()
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			var logEntry models.LogEntry
			if err := json.Unmarshal(message.Value, &logEntry); err != nil {
				h.logger.Errorf("Failed to unmarshal log entry: %v", err)
				continue
			}

			h.logChan <- logEntry
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}