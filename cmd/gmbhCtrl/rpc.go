package main

/**
 * handleRequest.go
 * Abe Dick
 * January 2019
 */

import (
	"context"
	"time"

	"github.com/gmbh-micro/cabal"
	"google.golang.org/grpc"
)

func getClient(address string) (cabal.ControlClient, context.Context, context.CancelFunc, error) {
	con, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, nil, nil, err
	}

	client := cabal.NewControlClient(con)
	if err != nil {
		return nil, nil, nil, err
	}

	ctx, can := context.WithTimeout(context.Background(), time.Second*2)
	return client, ctx, can, nil
}
