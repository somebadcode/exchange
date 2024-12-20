package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"os/signal"

	"github.com/somebadcode/exchange/converter"
)

type StatusCode = int

const (
	StatusOK StatusCode = iota
	StatusError
	StatusArgumentError
)

func main() {
	os.Exit(run())
}

func run() StatusCode {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	conv := &converter.Converter{
		AppID: os.Getenv("EXCHANGE_APP_ID"),
	}

	var qty *big.Rat
	var currency string

	flag.Func("qty", "quantity of the currency", func(s string) error {
		var err error
		qty, err = converter.ParseRat(s)
		if err != nil {
			return err
		}

		return nil
	})

	flag.StringVar(&conv.BaseCurrency, "base", "usd", "base currency")
	flag.StringVar(&currency, "currency", "eur", "currency to convert to")
	flag.Parse()

	if qty == nil {
		slog.Error("quantity is required (-qty)")

		return StatusArgumentError
	}

	if conv.AppID == "" {
		slog.Error("app id is required (environment variable `EXCHANGE_APP_ID`)")

		return StatusArgumentError
	}

	result, err := conv.ConvertTo(ctx, currency, qty)
	if err != nil {
		slog.Error("conversion failed",
			slog.String("baseCurrency", conv.BaseCurrency),
			slog.String("currency", currency),
			slog.String("qty", qty.String()),
			slog.String("op", "ConvertTo"),
			slog.String("err", err.Error()),
		)

		return StatusError
	}

	fmt.Printf("%s %s is %s %s\n", qty.FloatString(2), conv.BaseCurrency, result.FloatString(2), currency)

	return StatusOK
}
