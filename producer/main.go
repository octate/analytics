package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"
	"encoding/json"
	"github.com/Pallinder/go-randomdata"
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"

	"math/rand"
)

type UserLog struct {
	ServiceName	 string		`json:"ServiceName"`
	UserName	 string		`json:"UserName"`
	UserId		 string		`json:"UserId"`
	Location    	 string		`json:"Location"`
	ResponseCode	 int		`json:"ResponseCode"`
	Level 		 string		`json:"Level"`
	Instance	 string		`json:"Instance"`
	Timestamp 	 time.Time	`json:"timestamp"`
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


		// Setup configuration
		config := sarama.NewConfig()
		// The total number of times to retry sending a message (default 3).
		config.Producer.Retry.Max = 5
		// The level of acknowledgement reliability needed from the broker.
		config.Producer.RequiredAcks = sarama.WaitForAll
		brokers := []string{viper.GetString("kafka.server")}
		producer, err := sarama.NewAsyncProducer(brokers, config)
		if err != nil {
			panic(err)	 // Should not reach here
		}

		defer func() {
			if err := producer.Close(); err != nil {
				panic(err) 	// Should not reach here
			}
		}()

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		var enqueued, errors int
		doneCh := make(chan struct{})
		go func() {
			//pages := getPages()
			for i := 0; i < 100; i++ {

				time.Sleep(500 * time.Millisecond)
				uLog := getUserLog()
				cLog := getEndPointTest()

				//strTime := strconv.Itoa(int(time.Now().Unix()))
				msg1 := &sarama.ProducerMessage{
					Topic: viper.GetString("kafka.topic"),
					Key:   sarama.StringEncoder("userlog"),
					Value: sarama.StringEncoder(uLog.toString()),

				}
				msg2 := &sarama.ProducerMessage{
					Topic: viper.GetString("kafka.topic"),
					Key:   sarama.StringEncoder("EndPointTest"),
					Value: sarama.StringEncoder(cLog.toString()),

				}


				select {
				case producer.Input() <- msg1:
					enqueued++


					fmt.Println(msg1.Value)

				case producer.Input() <- msg2:
					enqueued++


					fmt.Println(msg2.Value)
				case err := <-producer.Errors():
					errors++
					fmt.Println("Failed to produce message:", err)
				case <-signals:
					doneCh <- struct{}{}
				}
				/*select {
				case producer.Input() <- msg2:
					enqueued++

					fmt.Println(msg2.Value)
				case err := <-producer.Errors():
					errors++
					fmt.Println("Failed to produce message:", err)
				case <-signals:
					doneCh <- struct{}{}
				}*/

			}
		}()

		<-doneCh
		log.Printf("Enqueued: %d; errors: %d\n", enqueued, errors)
	}
}


func getUserLog() UserLog{
	record:=UserLog{
			UserName:randomdata.LastName(),
			UserId:"blank",
			Timestamp:time.Now(),
			UserAction:userAction(),
			ResponseCode:200,
			Instance:randomdata.RandStringRunes(10),
			Level:"info",
			Location:randomdata.Country(randomdata.FullCountry),
			RequestId:randomdata.RandStringRunes(15),
			ServiceName:"qifAuth",
	}

	if(record.UserId=="blank"){
		record.UserId=getUserId(record.UserName)
	}
	return record


	}

func getEndPointTest() EndPointTest{
	log1 := EndPointTest{
		TestStatus:test(),
		Timestamp:time.Now(),
		ChannelName:randomdata.ChannelName(),
		TestId:randomdata.RandStringRunes(20),
	}
	return log1
}


func (p UserLog) toString() string {
	return toJson(p)
}
func (p EndPointTest) toString() string {
	return toJson(p)
}
func toJson(p interface{}) string {
	bytes, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return string(bytes)
}
func test() int {
	c := make(chan struct{})
	close(c)
	select {
	case <-c:
		return 0
	case <-c:
		return 1
	}
}
func getUserId(s string) string {
	name:=s
	var UserId string

	switch name {
	case "Smith":
		UserId="10001"
	case "Johnson":
		UserId="10002"
	case "Williams":
		UserId="10003"
	case "Jones":
		UserId="10004"
	case "Brown":
		UserId="10005"
	case "Davis":
		UserId="10006"
	case "Miller":
		UserId="10007"
	case "Wilson":
		UserId="10008"
	}
	return UserId

}
func userAction() string {
	rand.Seed(time.Now().Unix())
	reasons := []string{
		"LoginSuccess",
		"LogoutSuccess",
		"LogoutFail",
		"LoginFail",
	}
	n := rand.Int() % len(reasons)
	return reasons[n]
}