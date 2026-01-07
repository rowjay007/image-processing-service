package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"image-processing-service/internal/config"
	"image-processing-service/internal/ports"
)

type CloudAMQPQueue struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	cfg     config.CloudAMQPConfig
}

func NewCloudAMQPQueue(cfg config.CloudAMQPConfig) (*CloudAMQPQueue, error) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare queue to ensure it exists
	_, err = ch.QueueDeclare(
		cfg.QueueName,    // name
		cfg.QueueDurable, // durable
		false,            // delete when unused
		false,            // exclusive
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	if err := ch.Qos(
		cfg.PrefetchCount, // prefetch count
		0,                 // prefetch size
		false,             // global
	); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to set qos: %w", err)
	}

	return &CloudAMQPQueue{
		conn:    conn,
		channel: ch,
		cfg:     cfg,
	}, nil
}

func (q *CloudAMQPQueue) Publish(ctx context.Context, job *ports.TransformJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	err = q.channel.PublishWithContext(ctx,
		"",              // exchange
		q.cfg.QueueName, // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // Persistent if queue is durable
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	return nil
}

func (q *CloudAMQPQueue) Consume(ctx context.Context, handler func(*ports.TransformJob) error) error {
	msgs, err := q.channel.Consume(
		q.cfg.QueueName, // queue
		"",              // consumer
		false,           // auto-ack (FALSE = manual ack)
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			var job ports.TransformJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				log.Printf("Error unmarshalling job: %v", err)
				_ = d.Nack(false, false) // Dead letter or discard
				continue
			}

			// Execute handler
			if err := handler(&job); err != nil {
				log.Printf("Error processing job %s: %v", job.JobID, err)
				if nerr := d.Nack(false, true); nerr != nil {
					log.Printf("Error nacking message: %v", nerr)
				}
			} else {
				if aerr := d.Ack(false); aerr != nil {
					log.Printf("Error acking message: %v", aerr)
				}
			}
		}
	}()

	return nil
}

func (q *CloudAMQPQueue) Close() error {
	if q.channel != nil {
		_ = q.channel.Close()
	}
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}
