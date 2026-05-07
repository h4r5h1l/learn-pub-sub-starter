package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(val); err != nil {
		return err
	}
	return ch.PublishWithContext(context.Background(), exchange, key, false, false, amqp.Publishing{
		ContentType: "application/gob",
		Body:        buf.Bytes(),
	})
}

type SimpleQueueType string

const (
	Durable   SimpleQueueType = "durable"
	Transient SimpleQueueType = "transient"
)

type AckType string

const (
	Ack         AckType = "ack"
	NackRequeue AckType = "nackrequeue"
	NackDiscard AckType = "nackdiscard"
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	queueType SimpleQueueType,
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	var durable, autoDelete, exclusive bool
	switch queueType {
	case Durable:
		durable = true
		autoDelete = false
		exclusive = false
	case Transient:
		durable = false
		autoDelete = true
		exclusive = true
	}
	amqptable := map[string]any{"x-dead-letter-exchange": "peril_dlx"}
	queue, err := ch.QueueDeclare(queueName, durable, autoDelete, exclusive, false, amqptable)
	if err != nil {
		return nil, amqp.Queue{}, err
	}
	if err := ch.QueueBind(queue.Name, key, exchange, false, nil); err != nil {
		return nil, amqp.Queue{}, err
	}
	return ch, queue, nil
}

func SubscribeJSON[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	err = ch.Qos(10, 0, false)
	if err != nil {
		fmt.Printf("Error during Prefetch: %s", err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		defer ch.Close()
		for msg := range msgs {
			var val T
			if err := json.Unmarshal(msg.Body, &val); err != nil {
				msg.Nack(false, false)
				fmt.Printf("Negative Acknowledge occurred, Message Discarded\n")
				continue
			}
			acktype := handler(val)
			switch acktype {
			case Ack:
				msg.Ack(false)
				fmt.Printf("Acknowledge occurred\n")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Printf("Negative Acknowledge occurred, Message Requeued\n")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Printf("Negative Acknowledge occurred, Message Discarded\n")

			}
		}
	}()
	return nil
}

func SubscribeGob[T any](conn *amqp.Connection, exchange, queueName, key string, queueType SimpleQueueType, handler func(T) AckType) error {
	ch, q, err := DeclareAndBind(conn, exchange, queueName, key, queueType)
	if err != nil {
		return err
	}
	err = ch.Qos(10, 0, false)
	if err != nil {
		fmt.Printf("Error during Prefetch: %s", err)
	}
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		defer ch.Close()
		for msg := range msgs {
			buf := bytes.NewBuffer(msg.Body)
			dec := gob.NewDecoder(buf)
			var val T
			err := dec.Decode(&val)
			if err != nil {
				msg.Nack(false, false)
				fmt.Printf("Negative Acknowledge occurred, Message Discarded\n")
				continue
			}
			acktype := handler(val)
			switch acktype {
			case Ack:
				msg.Ack(false)
				fmt.Printf("Acknowledge occurred\n")
			case NackRequeue:
				msg.Nack(false, true)
				fmt.Printf("Negative Acknowledge occurred, Message Requeued\n")
			case NackDiscard:
				msg.Nack(false, false)
				fmt.Printf("Negative Acknowledge occurred, Message Discarded\n")

			}
		}
	}()
	return nil
}
