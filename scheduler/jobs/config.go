package jobs

import (
	"context"
	"errors"
	"fmt"
	elector "github.com/go-co-op/gocron-etcd-elector"
	"github.com/go-co-op/gocron/v2"
	"os"
	"time"
	_ "time/tzdata"
)

func Initialize() *JobsRepository {
	// Configuring elector with etdc cluster
	endpoint := os.Getenv("ETCD_HOST") + ":" + os.Getenv("ETCD_PORT")
	username := os.Getenv("ETCD_USERNAME")
	password := os.Getenv("ETCD_PASSWORD")
	cfg := elector.Config{
		Endpoints:   []string{endpoint},
		Username:    username,
		Password:    password,
		DialTimeout: 3 * time.Second,
	}

	// Build new elector
	el, err := elector.NewElector(context.Background(), cfg, elector.WithTTL(60))
	if err != nil {
		panic(err)
	}

	// el.Start() is a blocking method
	// so running with goroutine
	go func() {
		err := el.Start("/elections") // specify a directory for storing key value for election
		if errors.Is(err, elector.ErrClosed) {
			return
		}
	}()
	// New scheduler with elector
	TimeZone, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Time is: %v\n", time.Now().In(TimeZone))
	sh, err := gocron.NewScheduler(gocron.WithDistributedElector(el), gocron.WithLocation(TimeZone))
	if err != nil {
		panic(err)
	}
	sh.Start()
	return &JobsRepository{scheduler: &sh, elector: el}
}
