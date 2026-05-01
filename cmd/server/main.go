package main

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	fmt.Println("Starting Peril server...")
	conn_str := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(conn_str)
	if err != nil {
		fmt.Printf("Failed to connect to RabbitMQ: %s\n", err)
		return
	}
	rmqChannel, err := conn.Channel()
	if err != nil {
		fmt.Printf("Failed to create channel: %s\n", err)
		return
	}
	defer rmqChannel.Close()
	fmt.Println("Connected to RabbitMQ successfully!")

	//pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, "game_logs", "game_logs.*", pubsub.Durable)
	err = pubsub.SubscribeGob(conn, routing.ExchangePerilTopic, "game_logs", "game_logs.*", pubsub.Durable, handlerLogs())
	if err != nil {
		fmt.Printf("Failed to subscribe to game_logs: %s\n", err)
		return
	}

	gamelogic.PrintServerHelp()
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "pause":
			fmt.Println("Pausing game...")
			state := routing.PlayingState{IsPaused: true}
			pubsub.PublishJSON(rmqChannel, routing.ExchangePerilDirect, routing.PauseKey, state)
		case "resume":
			fmt.Println("Resuming game...")
			state := routing.PlayingState{IsPaused: false}
			pubsub.PublishJSON(rmqChannel, routing.ExchangePerilDirect, routing.PauseKey, state)
		case "quit":
			fmt.Println("Quitting game...")
			return
		default:
			fmt.Printf("Unknown command: %s\n", input[0])
		}
	}
}
