package db

import "gohighlevel/pkg/types"

// Database interface defines the methods that any database implementation must provide
type Database interface {
	Init() error
	Close()
	InsertSample(sample types.Sample) error
}
