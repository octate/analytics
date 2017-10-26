package main

import (
	"fmt"
	"os"
	"os/signal"
	"github.com/Shopify/sarama"
	"gopkg.in/mgo.v2"
	//"testing"
	//"gopkg.in/mgo.v2/bson"
	//"time"
	//"gopkg.in/mgo.v2/bson"
	//"time"
	//"encoding/json"
	"encoding/json"
	//"gopkg.in/mgo.v2/bson"
	//"gopkg.in/mgo.v2/bson"
	"time"
	"github.com/spf13/viper"
)
type UserLog struct {
	ServiceName	 string		`json:"ServiceName"`
	UserName	 string		`json:"UserName"`
	UserId		 string		`json:"UserId"`
	Location    	 string		`json:"Location"`
	ResponseCode	 int		`json:"ResponseCode"`
	Level 		 string		`json:"Level"`
	Instance	 string		`json:"Instance"`
	Timestamp 	 time.Time	`json:"Timestamp"`
	RequestId	 string		`json:"RequestId"`
	UserAction	 string		`json:"UserAction"`
}

type EndPointTest struct {
	TestStatus 	   int  	  `json:"test_status"`
	ChannelName 	   string  	  `json:"channel_name"`
	Timestamp   	   time.Time  	  `json:"timestamp"`
	TestId		   string	  `json:"test_id"`
}

func main() {

	viper.SetConfigName("app")
	viper.AddConfigPath("src/config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	} else {
		//dev_server := viper.GetString("mongo.server")

		session, err := mgo.Dial(viper.GetString("mongo.server"))
		if err != nil {
			panic(err)
		}

		//defer session.Close()

		session.SetMode(mgo.Monotonic, true)




		// Collection People
		userLogCollection := session.DB(viper.GetString("mongo.database")).C(viper.GetString("mongo.user_activity"))
		channelTestCollection := session.DB(viper.GetString("mongo.database")).C(viper.GetString("mongo.end_point_tests"))
		// Index
		index := mgo.Index{
			Key:        []string{"RequestId"},
			Unique:     true,
			DropDups:   true,
			Background: true,
			Sparse:     true,
		}
		err = userLogCollection.EnsureIndex(index)
		if err != nil {
			panic(err)
		}


		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true

		// Specify brokers address. This is default one
		brokers := []string{viper.GetString("kafka.server")}

		// Create new consumer
		master, err := sarama.NewConsumer(brokers, config)
		if err != nil {
			panic(err)
		}

		defer func() {
			if err := master.Close(); err != nil {
				panic(err)
			}
		}()

		topic := viper.GetString("kafka.topic")
		// How to decide partition, is it fixed value...?
		consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		// Count how many message processed
		msgCount := 0

		// Get signnal for finish
		doneCh := make(chan struct{})
		go func() {
			for {
				select {
				case err := <-consumer.Errors():
					fmt.Println(err)
				case msg := <-consumer.Messages():
					msgCount++

					key := getUserLog(msg.Value).ServiceName

					if (key == "qifAuth") {
						record := getUserLog(msg.Value)
						err = userLogCollection.Insert(record)
						fmt.Println("userLog messages", record)
					}else {
						record := getEndPointLog(msg.Value)
						err = channelTestCollection.Insert(record)
						fmt.Println("endPointTest messages", record)
					}

				case <-signals:
					fmt.Println("Interrupt is detected")
					doneCh <- struct{}{}
				}
			}

		}()

		<-doneCh
		fmt.Println("Processed", msgCount, "messages")
	}
}

func getUserLog(recievedMsg []byte) UserLog {
	raw := recievedMsg
	var temp UserLog
	json.Unmarshal(raw, &temp)
	return temp
}
func getEndPointLog(recievedMsg []byte) EndPointTest {
	raw := recievedMsg
	var temp EndPointTest
	json.Unmarshal(raw, &temp)
	return temp
}