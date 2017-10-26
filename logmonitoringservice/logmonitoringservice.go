package main

import (
	"fmt"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
	"encoding/json"
	"time"
	"strconv"

	"github.com/spf13/viper"
)

//gives list of successful logins grouped per day
func userlog(ctx *fasthttp.RequestCtx) {
	viper.SetConfigName("app")
	viper.AddConfigPath("src/config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	} else {
		session, err := mgo.Dial(viper.GetString("mongo.server"))
		if err != nil {
			panic(err)
		}

		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(viper.GetString("mongo.database")).C(viper.GetString("mongo.user_activity"))

		value:= ctx.UserValue("days")


		data := value.(string)
		i, err := strconv.Atoi(data)
		if err != nil {
			fmt.Fprint(ctx, "Invalid number of days : ", value )
		} else {
			i = i * (-1)
			current_time := time.Now()
			from := current_time.AddDate(0, 0, i)

			o1 := bson.M{
				"$match" :bson.M{"useraction":"LoginSuccess"},						//filter records where useraction is "LoginSuccess"
			}
			o2 := bson.M{
				"$match" :bson.M{
					"timestamp": bson.M{								//filter records where timestamp is greater than "from" timestamp
						"$gt": from,

					},
				},
			}
			o3 := bson.M{											//groups the data for each date derived from the timestamp
				"$group": bson.M{
					"_id": bson.M{
						"year": bson.M{"$year": "$timestamp" },
						"dayOfYear": bson.M{"$dayOfYear": "$timestamp" },
					},
					"success_login": bson.M{"$sum": 1 },						//counts number of total records
					"date": bson.M{"$first": "$timestamp"},						//stores first timestamp from the group
				},
			}
			o4 := bson.M{
				"$project": bson.M{									//formats data to present in the response
					"_id": 0,
					"success_login_count": "$success_login",
					"date": bson.M{
						"$dateToString": bson.M{"format": "%Y-%m-%d", "date": "$date" },	//converts timestamp to "Y-m-d" format
					},
					"count": 1,
				},
			}

			operations := []bson.M{o1, o2, o3, o4}
			pipe := c.Pipe(operations)
			results := []bson.M{}
			err1 := pipe.All(&results)

			if err1 != nil {
				fmt.Printf("ERROR : %s\n", err1.Error())
				return
			}
			b, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Fprint(ctx, string(b))
		}


	}

}

//List of users for the last "x" days
func userlist(ctx *fasthttp.RequestCtx) {
	viper.SetConfigName("app")
	viper.AddConfigPath("src/config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	} else {
		session, err := mgo.Dial(viper.GetString("mongo.server"))
		if err != nil {
			panic(err)
		}

		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(viper.GetString("mongo.database")).C(viper.GetString("mongo.user_activity"))

		value := ctx.UserValue("days")
		data := value.(string)
		i, err := strconv.Atoi(data)
		if err != nil {
			fmt.Fprint(ctx, "Invalid number of days : ", value )
		} else {
			i = i * (-1)

			current_time := time.Now()
			from := current_time.AddDate(0, 0, i)

			o1 := bson.M{
				"$match" :bson.M{"useraction":"LoginSuccess"},						//filter records where useraction is "LoginSuccess"
			}
			o2 := bson.M{
				"$match" :bson.M{									//filter records where timestamp is greater than "from" timestamp
					"timestamp": bson.M{
						"$gt": from,
					},
				},
			}
			o3 := bson.M{
				"$group" :bson.M{
					"_id": "$username",								//groups the data with respect for each userid
					"last_login_timestamp" :  bson.M{
						"$last":"$timestamp",							//returns the last timestamp
					},
					"user_id": bson.M{
						"$last":"$userid",							//returns the last timestamp
					},
				},

			}
			o4:=bson.M{

				"$project":bson.M{									//formats data to present in the response
					"_id": 0,
					"user_name":"$_id",								//stores the value of "_id" from previous stage to "user_id"
					"user_id":"$user_id",
					"last_login_timestamp" : "$last_login_timestamp",				//stores the value of "last_login_timestamp" from previous stage to "last_login_timestamp"
				},
			}


			operations := []bson.M{o1, o2, o3, o4}
			pipe := c.Pipe(operations)
			results := []bson.M{}
			err1 := pipe.All(&results)

			if err1 != nil {
				fmt.Printf("ERROR : %s\n", err1.Error())
				return
			}
			b, err := json.MarshalIndent(results, "", "  ")
			if err != nil {
				fmt.Println("error:", err)
			}
			fmt.Fprint(ctx, string(b))
		}
	}
}

// endtest is the channel endtest handler
func endtest(ctx *fasthttp.RequestCtx) {
	viper.SetConfigName("app")
	viper.AddConfigPath("src/config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	} else {
		session, err := mgo.Dial(viper.GetString("mongo.server"))
		if err != nil {
			panic(err)
		}
		defer session.Close()
		session.SetMode(mgo.Monotonic, true)
		c := session.DB(viper.GetString("mongo.database")).C(viper.GetString("mongo.end_point_tests"))
		o1 := bson.M{
			"$project" :bson.M{										//Feeding only channelname, teststatus and timestamp to the next stage
				"channel": "$channelname",
				"status" : "$teststatus",
				"time" : "$timestamp",
				"_id":1,
			},
		}
		o2 := bson.M{
			"$group" :bson.M{
				"_id": "$channel",									//grouping the day for each channelname
				"Success": bson.M{
					"$sum": "$status"},								//calculating the total passed endpoint tests
				"Total": bson.M{									//calculating total number of endtests for each channel
					"$sum": 1},
				"last_status": bson.M{
					"$last":bson.M{									//Getting the last teststatus of each channel
						"$cond": []interface{}{
							bson.M{"$eq": []interface{}{"$status", 0}},
							"fail", "paas",
						},
					},
				},
				"last_status_timestamp" : bson.M{							//Getting the timestamp of last teststatus for each channel
					"$last":"$time",
				},
			},
		}

		o3 := bson.M{
			"$project" :bson.M{
				"channel_name": "$_id",
				"passed_tests_count":"$Success",
				"total_tests_count" : "$Total",
				"_id":0,
				"last_test_status" : "$last_status",
				"last_test_status_timestamp":"$last_status_timestamp",
				"failed_tests_count": bson.M{								//calculating total number of failed endtests for each channel by calculating (total paased - total failed)
					"$subtract": []interface{}{"$Total", "$Success"},
				},
			},
		}

		operations := []bson.M{o1, o2, o3}
		pipe := c.Pipe(operations)
		results := []bson.M{}
		err1 := pipe.All(&results)

		if err1 != nil {
			fmt.Printf("ERROR : %s\n", err1.Error())
			return
		}
		b, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		fmt.Fprint(ctx, string(b))
	}
}

func main() {
	fmt.Println("Server Started")
	router := fasthttprouter.New()
	router.GET("/login_stats/:days", userlog)
	router.GET("/users/:days", userlist)
	router.GET("/channel_endpoint_tests/summary/", endtest)
	log.Fatal(fasthttp.ListenAndServe("localhost:8082", router.Handler))
}
