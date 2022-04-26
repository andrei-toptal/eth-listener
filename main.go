package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	log.Println("Starting eth-listener application...")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

		sig := <-ch
		log.Printf("Shutting down due to %s", sig)
		cancel()
	}()

	app, err := WireApp(ConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	defer app.tokensDB.Close()

	log.Println("Watching for transactions...")

	headsCh := make(chan *types.Header)
	sub, err := app.client.SubscribeNewHead(ctx, headsCh)
	if err != nil {
		log.Fatal(err)
	}
	defer sub.Unsubscribe()

	transfersCh := make(chan *Transfer, TransfersChBuffer)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case transfer := <-transfersCh:
				handleTransfer(transfer, app, ctx)
			}
		}
	}()

mainLoop:
	for {
		select {
		case <-ctx.Done():
			break mainLoop
		case err := <-sub.Err():
			log.Fatalf("Head subscription error: %v", err)
		case header := <-headsCh:
			handleHeader(ctx, header, transfersCh, app)
		}
	}

	app.telegram.Notify("Bot is shutting down...")
	log.Printf("Application stopped.")
}
