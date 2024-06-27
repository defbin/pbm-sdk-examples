package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/percona/percona-backup-mongodb/sdk"
)

const DefaultMongoURI = "mongodb://localhost:27017"

func main() {
	ctx := context.Background()

	uri := os.Getenv("DATABASE_URI")
	if uri == "" {
		uri = DefaultMongoURI
	}

	pbm, err := sdk.NewClient(ctx, uri)
	if err != nil {
		log.Fatalf("new sdk client: %v", err)
	}
	defer func() {
		err := pbm.Close(context.Background())
		if err != nil {
			log.Printf("close sdk client: %v", err)
		}
	}()

	locks, err := pbm.OpLocks(ctx)
	if err != nil {
		log.Fatalf("oplock: %v", err)
	}

	if len(locks) == 0 {
		fmt.Println("<noop>")
		return
	}

	for _, lock := range locks {
		fmt.Printf("> cmd: %s; rs: %q; node: %q",
			lock.Cmd,
			lock.Replset,
			lock.Node)
		if err := lock.Err(); err != nil {
			if errors.Is(err, sdk.ErrStaleHearbeat) {
				fmt.Printf(" stuck: " + err.Error())
			} else {
				fmt.Printf(" ERROR: " + err.Error())
			}
		}
		fmt.Println()
	}
}
