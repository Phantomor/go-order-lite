package mq

import (
	"go-order-lite/pkg/config"
	"go-order-lite/pkg/logger"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
)

var Producer rocketmq.Producer

func InitMQ() error {
	var err error
	Producer, err = rocketmq.NewProducer(
		producer.WithNameServer(config.Cfg.RocketMQ.NameServers),
		producer.WithRetry(config.Cfg.RocketMQ.Retry),
		producer.WithGroupName("go-order-lite-producter-group"),
	)
	if err != nil {
		return err
	}
	if err = Producer.Start(); err != nil {
		return err
	}

	logger.Log.Info("RocketMQ Producer started successfully",
		zap.Strings("name_servers", config.Cfg.RocketMQ.NameServers),
	)
	return nil
}

func CloseMQ() error {
	if Producer != nil {
		err := Producer.Shutdown()
		if err != nil {
			logger.Log.Error("Failed to shutdown RocketMQ Producer", zap.Error(err))
			return err
		}
		logger.Log.Info("RocketMQ Producer shutdown successfully")
	}
	return nil
}
