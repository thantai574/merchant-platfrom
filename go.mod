module orders-system

go 1.14

replace google.golang.org/grpc v1.29.1 => google.golang.org/grpc v1.26.0

require (
	github.com/Shopify/sarama v1.27.2
	github.com/dustin/go-humanize v1.0.0
	github.com/eclipse/paho.mqtt.golang v1.3.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api v4.6.4+incompatible
	github.com/golang/protobuf v1.4.0
	github.com/grpc-ecosystem/go-grpc-middleware v1.1.0
	github.com/jakehl/goid v1.1.0
	github.com/leekchan/accounting v0.3.1
	github.com/lysu/kazoo-go v0.0.0-20160229162054-26744d16cdcf
	github.com/micro/go-micro/v2 v2.9.1
	github.com/panjf2000/ants/v2 v2.4.3
	github.com/samuel/go-zookeeper v0.0.0-20201211165307-7117e9ea2414 // indirect
	github.com/spf13/cast v1.3.0
	github.com/spf13/viper v1.7.1
	github.com/streadway/amqp v1.0.0
	github.com/stretchr/testify v1.6.1
	go.mongodb.org/mongo-driver v1.4.3
	go.uber.org/zap v1.16.0
	golang.org/x/mod v0.1.1-0.20191107180719-034126e5016b // indirect
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/sys v0.0.0-20201231184435-2d18734c6014 // indirect
	golang.org/x/tools v0.0.0-20200207183749-b753a1ba74fa // indirect
	google.golang.org/grpc v1.26.0
	google.golang.org/protobuf v1.26.0-rc.1
	gopkg.in/maddevsio/fcm.v1 v1.0.5
)
