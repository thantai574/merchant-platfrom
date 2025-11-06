package kafka

import (
	"context"
	"testing"
)

func TestNewConnectionConsumer(t *testing.T) {
	type args struct {
		ctx        context.Context
		brokers    string
		zookeepers string
		group      string
		topics     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy_case",
			args: args{
				ctx:        context.Background(),
				brokers:    "35.240.196.102:9094",
				zookeepers: "34.87.84.225:2181",
				group:      "group-test",
				topics:     "112",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, err := NewConnection(context.TODO(), tt.args.zookeepers, tt.args.brokers)
			if err != nil {
				t.Errorf("err %s ", err)
			}

		})
	}
}
