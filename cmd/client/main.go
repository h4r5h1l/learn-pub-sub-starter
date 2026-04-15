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
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Printf("Failed to welcome client: %s\n", err)
		return
	}
	pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+username, routing.PauseKey, pubsub.Transient)
	gamestate := gamelogic.NewGameState(username)
	for {
		input := gamelogic.GetInput()
		if len(input) == 0 {
			continue
		}
		switch input[0] {
		case "spawn": // Example usage: spawn europe infantry
			fmt.Println("Spawning unit...")
			err := gamestate.CommandSpawn(input)
			if err != nil {
				fmt.Printf("Failed to spawn unit: %s\n", err)
			}
		case "move": //Example usage: move europe 1
			fmt.Println("Moving unit...")
			_, err := gamestate.CommandMove(input)
			if err != nil {
				fmt.Printf("Failed to move unit: %s\n", err)
			}
			fmt.Printf("Successfully moved unit!\n")
		case "status":
			fmt.Println("Checking game status...")
			gamestate.CommandStatus()
		case "help":
			gamelogic.PrintClientHelp()
		case "spam":
			fmt.Println("Spamming not allowed yet!")
		case "quit":
			gamelogic.PrintQuit()
			return
		default:
			fmt.Printf("Unknown command: %s\n", input[0])
		}
	}
}
