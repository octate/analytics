# analytics
Log monitoring service POC
# Pre-requisites:
 -GO - 1.8
 -mongodb,
 -kafka
# How to test:
1. Update the config/config.toml file as per your system
2. Update the path of config file in logmonitoringservice.go,producer.go and consumer.go
 - Run consumer.go (It listens on the particular topic and puts the logs in mongodb).
 - Run Producer.go (It generates dummy logs and puts on a particular topic in Kafka. You can change the topic name in config file).
 - Run logmonitoringservice.go (It has the bussiness logic and exposes ReST apis through which you can view some analytics).
