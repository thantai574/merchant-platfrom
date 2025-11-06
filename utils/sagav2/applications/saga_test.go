package applications

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"orders-system/infrastructure/kafka"
	"orders-system/utils/configs"
	"orders-system/utils/gpooling"
	"orders-system/utils/sagav2/domains"
	"testing"
	"time"
)

func TestNewCoordinator(t *testing.T) {
	config, err := configs.LoadTestConfig("../../../")
	if err != nil {
		t.Error(err)
	}
	pool_go_routine, _ := gpooling.NewPooling(config.MaxPoolSize)

	type args struct {
		id   string
		ctx  context.Context
		pool gpooling.IPool
		conf *configs.Config
	}

	tests := []struct {
		name    string
		args    args
		want    *Saga
		wantErr bool
		error1  error
		error2  error
	}{
		{
			name: "none error ",
		},
		{
			name:   "error step 1 ",
			error1: fmt.Errorf("1"),
		},

		{
			name:   "error step 2 ",
			error2: fmt.Errorf("2"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancelFN := context.WithTimeout(context.Background(), time.Second*10)
			kafka_conn, err := kafka.NewConnection(context.Background(), config.KafkaConfig.Zookeepers, config.KafkaConfig.Brokers)
			if err != nil {
				if err != nil {
					t.Error(err)
				}
			}
			got, err := NewCoordinator(fmt.Sprint(time.Now().Unix()), ctx, pool_go_routine, config, kafka_conn)

			compensate1 := false
			compensate2 := false

			res := got.WithStep(&domains.Step{
				Name: "Step_1",
				Func: func(ctx context.Context) error {
					fmt.Println("step 1")
					return tt.error1
				},
				CompensateFunc: func(ctx context.Context) error {
					fmt.Println("Cancel step 1")
					compensate1 = true
					return nil
				},
			}).WithStep(&domains.Step{
				Name: "Step_2",
				Func: func(ctx context.Context) error {
					fmt.Println("step 2")
					return tt.error2
				},
				CompensateFunc: func(ctx context.Context) error {
					fmt.Println("Cancel step 2")
					compensate2 = true
					return nil
				},
			}).Play()
			cancelFN()

			if tt.error2 != nil {
				assert.Equal(t, true, compensate2)
				assert.Equal(t, true, compensate1)
			}
			if tt.error1 != nil {
				assert.Equal(t, true, compensate1)
			}

			if tt.error2 == nil && tt.error1 == nil {
				assert.Equal(t, false, compensate1)
				assert.Equal(t, false, compensate2)
				assert.Equal(t, nil, res.ExecutionError)
			}

			return
		})
	}
}
