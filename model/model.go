package model

import (
	"context"

	"simple-agent/types"
)

type Model interface {
	NewStreaming(ctx context.Context, history []types.Message, question string) Stream
}

type Stream interface {
	Next() bool
	Err() error
	Current() string
	GetResponse() string
}

type modelImpl struct {
	modelName string
	maxToken  int
}
