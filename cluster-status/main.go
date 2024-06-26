package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/percona/percona-backup-mongodb/sdk"
	"github.com/percona/percona-backup-mongodb/sdk/cli"
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

	clusterStatus, err := cli.ClusterStatus(ctx, pbm, cli.RSConfGetter(uri))
	if err != nil {
		log.Fatalf("cluster status: %v", err)
	}

	for replsetName, nodes := range clusterStatus {
		fmt.Printf("replset: %q\n", replsetName)

		for _, node := range nodes {
			switch {
			case node.Ver == "":
				fmt.Printf("- %s [n/a]\n", node.Host)

			case node.OK:
				fmt.Printf("- %s (%s) [PBM %s] OK\n",
					node.Host,
					node.Role,
					node.Ver)

			case node.IsAgentLost():
				fmt.Printf("- %s (%s) [PBM %s] ERROR\n",
					node.Host,
					node.Role,
					node.Ver)

			default: // !node.OK
				fmt.Printf("- %s (%s) [PBM %s] ERROR\n",
					node.Host,
					node.Role,
					node.Ver)
			}

			for _, err := range node.Errs {
				fmt.Printf("  | %s\n", err)
			}
		}

		fmt.Println()
	}
}
