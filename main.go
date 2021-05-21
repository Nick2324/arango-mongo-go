package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ArangoResults struct {
	Key    string `json:"_key"`
	Id     string `json:"_id"`
	Rev    string `json:"_rev"`
	Nombre string `json:"nombre"`
	Tipo   string `json:"tipo"`
	Etapa  string `json:"etapa"`
}

type RutaIdMonto struct {
	Id string `json:"_id"`
}

type RutaMongo struct {
	Id   RutaIdMonto      `json:"id"`
	Path []ActividadMongo `json:"path"`
}

type ActividadMongo struct {
	Id     string `json:"_id"`
	Nombre string `json:"nombre"`
	Pool   string `json:"pool"`
	Tipo   string `json:"tipo"`
	Flow   string `json:"flow"`
}

type FlujoMongo struct {
}

func getDataWithArango() {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://localhost:8529"},
	})
	if err != nil {
		panic(err)
	}
	c, err := driver.NewClient(driver.ClientConfig{
		Connection:     conn,
		Authentication: driver.BasicAuthentication("root", "mifamilia"),
	})
	if err != nil {
		panic(err)
	}
	db, err := c.Database(context.Background(), "workflow")
	if err != nil {
		panic(err)
	}
	cursor, err := db.Query(context.Background(),
		"FOR v, e, p IN 0..99 OUTBOUND @eventId GRAPH @graph  PRUNE (v._id != @eventId && !IS_SAME_COLLECTION(@collectionName, v._id)) RETURN v",
		map[string]interface{}{
			"eventId":        "event/33797",
			"graph":          "selsa_wf",
			"collectionName": "activity",
		})
	if err != nil {
		panic(err)
	}
	result := &ArangoResults{}
	for {
		_, err := cursor.ReadDocument(context.Background(), result)
		if err != nil {
			if err.Error() != "no more documents" {
				panic(err)
			} else {
				break
			}
		}
		if result == nil {
			break
		}
		fmt.Printf("%v\n", result)
	}

}

func getDataWithMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://127.0.0.1:27017/?ssl=false"))
	if err != nil {
		panic(err)
	}
	db := client.Database("workflow")
	cursor, err := db.Collection("event").Aggregate(ctx, mongo.Pipeline{
		bson.D{{"$match", bson.D{{"etapa", "inicial"}}}},
		bson.D{{"$graphLookup", bson.D{{"from", "activity"}, {"startWith", "$flow.actividad"}, {"connectFromField", "flow.actividad"}, {"connectToField", "_id"}, {"as", "path"}, {"depthField", "depth"}}}},
		bson.D{{"$unwind", bson.D{{"path", "$path"}}}},
		bson.D{{"$sort", bson.D{{"depth", -1}}}},
		bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "$_id"}}}, {"path", bson.D{{"$push", "path"}}}}}},
	})
	if err != nil {
		panic(err)
	}

	result := map[string]interface{}{}
	cursor.All(context.Background(), result)
	fmt.Printf("%v\n", result)
}

func main() {
	//getDataWithMongo()
	//getDataWithArango()
	pruebas()
}

func pruebas() {
	var k *int
	fmt.Printf("%v\n", reflect.ValueOf(k).IsNil())
	j := 1
	i := &j
	fmt.Println(*i)
	cambiarValor(i)
	fmt.Println(*i)
	fmt.Println(i)
}

func cambiarValor(value interface{}) {
	var i interface{}
	j := 2
	i = &j
	val1 := reflect.ValueOf(i)
	val := reflect.ValueOf(value).Elem()
	fmt.Println(val.Interface())
	//val.Set(reflect.ValueOf(*i))
	val.Set(reflect.ValueOf(val1.Elem().Interface()))
}
