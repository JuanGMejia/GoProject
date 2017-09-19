/*package main

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

// Tweet is a structure used for serializing/deserializing data in Elasticsearch.
type Tweet struct {
	User     string                `json:"user"`
	Message  string                `json:"message"`
	Retweets int                   `json:"retweets"`
	Image    string                `json:"image,omitempty"`
	Created  time.Time             `json:"created,omitempty"`
	Tags     []string              `json:"tags,omitempty"`
	Location string                `json:"location,omitempty"`
	Suggest  *elastic.SuggestField `json:"suggest_field,omitempty"`
}

const mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"tweet":{
			"properties":{
				"user":{
					"type":"keyword"
				},
				"message":{
					"type":"text",
					"store": true,
					"fielddata": true
				},
				"image":{
					"type":"keyword"
				},
				"created":{
					"type":"date"
				},
				"tags":{
					"type":"keyword"
				},
				"location":{
					"type":"geo_point"
				},
				"suggest_field":{
					"type":"completion"
				}
			}
		}
	}
}`

func main() {
	// Starting with elastic.v5, you must pass a context to execute each service
	ctx := context.Background()

	// Obtain a client and connect to the default Elasticsearch installation
	// on 127.0.0.1:9200. Of course you can configure your client to connect
	// to other hosts and configure it in various other ways.
	client, err := elastic.NewClient()
	if err != nil {
		// Handle error
		panic(err)
	}

	// Ping the Elasticsearch server to get e.g. the version number
	info, code, err := client.Ping("http://127.0.0.1:9200").Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)

	// Getting the ES version number is quite common, so there's a shortcut
	esversion, err := client.ElasticsearchVersion("http://127.0.0.1:9200")
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Elasticsearch version %s\n", esversion)

	// Use the IndexExists service to check if a specified index exists.
	exists, err := client.IndexExists("twitter").Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if !exists {
		// Create a new index.
		createIndex, err := client.CreateIndex("twitter").BodyString(mapping).Do(ctx)
		if err != nil {
			// Handle error
			panic(err)
		}
		if !createIndex.Acknowledged {
			// Not acknowledged
		}
	}

	// Index a tweet (using JSON serialization)

	tweet1 := Tweet{User: "olivere", Message: "Take Five", Retweets: 0}
	put1, err := client.Index().
		Index("twitter").
		Type("tweet").
		Id("1").
		BodyJson(tweet1).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index %s, type %s\n", put1.Id, put1.Index, put1.Type)

	// Index a second tweet (by string)
	tweet2 := `{"user" : "olivere", "message" : "It's a Raggy Waltz"}`
	put2, err := client.Index().
		Index("twitter").
		Type("tweet").
		Id("2").
		BodyString(tweet2).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("Indexed tweet %s to index %s, type %s\n", put2.Id, put2.Index, put2.Type)

	// Get tweet with specified ID
	get1, err := client.Get().
		Index("twitter").
		Type("tweet").
		Id("1").
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if get1.Found {
		fmt.Printf("Got document %s in version %d from index %s, type %s\n", get1.Id, get1.Version, get1.Index, get1.Type)
	}

	// Flush to make sure the documents got written.
	_, err = client.Flush().Index("twitter").Do(ctx)
	if err != nil {
		panic(err)
	}

	// Search with a term query
	termQuery := elastic.NewTermQuery("user", "olivere")
	searchResult, err := client.Search().
		Index("twitter").   // search in index "twitter"
		Query(termQuery).   // specify the query
		Sort("user", true). // sort by "user" field, ascending
		From(0).Size(10).   // take documents 0-9
		Pretty(true).       // pretty print request and response JSON
		Do(ctx)             // execute
	if err != nil {
		// Handle error
		panic(err)
	}

	// searchResult is of type SearchResult and returns hits, suggestions,
	// and all kinds of other information from Elasticsearch.
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)

	// Each is a convenience function that iterates over hits in a search result.
	// It makes sure you don't need to check for nil values in the response.
	// However, it ignores errors in serialization. If you want full control
	// over iterating the hits, see below.
	var ttyp Tweet
	for _, item := range searchResult.Each(reflect.TypeOf(ttyp)) {
		if t, ok := item.(Tweet); ok {
			fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
		}
	}
	// TotalHits is another convenience function that works even when something goes wrong.
	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())

	// Here's how you iterate through results with full control over each step.
	if searchResult.Hits.TotalHits > 0 {
		fmt.Printf("Found a total of %d tweets\n", searchResult.Hits.TotalHits)

		// Iterate through results
		for _, hit := range searchResult.Hits.Hits {
			// hit.Index contains the name of the index

			// Deserialize hit.Source into a Tweet (could also be just a map[string]interface{}).
			var t Tweet
			err := json.Unmarshal(*hit.Source, &t)
			if err != nil {
				// Deserialization failed
			}

			// Work with tweet
			fmt.Printf("Tweet by %s: %s\n", t.User, t.Message)
		}
	} else {
		// No hits
		fmt.Print("Found no tweets\n")
	}

	// Update a tweet by the update API of Elasticsearch.
	// We just increment the number of retweets.
	update, err := client.Update().Index("twitter").Type("tweet").Id("1").
		Script(elastic.NewScriptInline("ctx._source.retweets += params.num").Lang("painless").Param("num", 1)).
		Upsert(map[string]interface{}{"retweets": 0}).
		Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	fmt.Printf("New version of tweet %q is now %d\n", update.Id, update.Version)

	// ...

	// Delete an index.
	deleteIndex, err := client.DeleteIndex("twitter").Do(ctx)
	if err != nil {
		// Handle error
		panic(err)
	}
	if !deleteIndex.Acknowledged {
		// Not acknowledged
	}
}*/
package main

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"database/sql"
)

var(
	db *gorm.DB
	err error
)


type Customer struct {
	CustomerID int `gorm:"column:CustomerID"`
	CustomerName string `gorm:"column:CustomerName"`
	ContactName string `gorm:"column:ContactName"`
	Country string `gorm:"column:Country"`
}

type Order struct {
	OrderID int `gorm:"column:OrderID"`
	CustomerID int `gorm:"column:CustomerID"`
	OrderDate time.Time `gorm:"column:OrderDate"`
}

func (c *Customer) Create() error{
	create:=db.Create(&c)
	fmt.Println(create)
	if create.RowsAffected != 0 {
		fmt.Println("El usuario fue creado")
	}else{
		return errors.New("No se pudo crear")
	}
	return nil
}

func (o *Order) Create() error{
	create:=db.Create(&o)
	fmt.Println(create)
	if create.RowsAffected != 0 {
		fmt.Println("La orden fue creada")
	}else{
		return errors.New("No se pudo crear la orden")
	}
	return nil
}



func init(){
	db, err = gorm.Open("mysql", "root:root@/db?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}

}

func main() {

	defer db.Close()

	//var orders []interface{}
	//PopulateCustomers()
	//PopulateOrders()
	//var interfaceM []interface{}
	//var Resultinterface []Order
	var rows *sql.Rows
	var customer Customer
	var order Order
	rows,err = db.Raw(`SELECT c.CustomerName,o.OrderDate
FROM Customers c
INNER JOIN Orders o ON o.CustomerID=c.CustomerID
WHERE o.OrderID <= 10308`).Rows()
	//rows,err = db.Table("customers c").Select("c.CustomerName, o.OrderID").Joins("inner join orders o on c.CustomerID=o.CustomerID").Where("o.OrderID <= 10308").Rows()
	defer rows.Close()
	//fmt.Printf("(%v, %T)\n", rows, rows)
	for rows.Next() {
	//rows.Scan()
	//var order interface{}
		//rows.Scan(&order)
		// var x string
		rows.Scan(&customer.CustomerName,&order.OrderDate)
		//db.ScanRows(rows,&order)
		fmt.Println(customer.CustomerName)
		fmt.Printf("%d-%02d-%02d",order.OrderDate.Year(),order.OrderDate.Month(),order.OrderDate.Day())
		//fmt.Printf("(%v)\n", order)



	}


	//var customers []Customer
	//db.Where("CustomerID = ?", 2).First(&customers)
	/*db.Find(&customers)
	for i:=1; i<=len(customers);i++{
		fmt.Println("El cliente ",i," es:")
		fmt.Println(customers[i-1].CustomerID)
		fmt.Println(customers[i-1].CustomerName)
		fmt.Println(customers[i-1].ContactName)
		fmt.Println(customers[i-1].Country)
	}*/

	/*
	var users interface{}
	users = db.Find(&user)
	users2 := users.(*gorm.DB)
	fmt.Printf("(%v, %T)\n", users2, users2)

	// Migrate the schema
	db.AutoMigrate(&User{})

	// Create
	db.Create(&User{})

	// Read
	var product Product
	db.First(&product, 1) // find product with id 1
	db.First(&product, "code = ?", "L1212") // find product with code l1212

	// Update - update product's price to 2000
	db.Model(&product).Update("Price", 2000)

	// Delete - delete product
	db.Delete(&product)
	*/

}

func PopulateCustomers() {
	customer := Customer{1,"Alfreds Futterkiste","Maria Anders","Germany"}
	customer1 := Customer{2,"Ana Trujillo Emparedados y helados","Ana Trujillo","Mexico"}
	customer2 := Customer{3,"Antonio Moreno Taquería","Antonio Moreno","Mexico"}
	err = customer.Create()
	if err!=nil {
		panic(err)
	}
	err = customer1.Create()
	if err!=nil {
		panic(err)
	}
	err = customer2.Create()
	if err!=nil {
		panic(err)
	}
}

func PopulateOrders() {

	order := Order{10308,2,time.Date(1996,time.Month(9),18, 0, 0, 0, 0, time.UTC)}
	order1 := Order{10309,37,time.Date(1996,time.Month(9),19, 0, 0, 0, 0, time.UTC)}
	order2 := Order{10310,77,time.Date(1996,time.Month(9),20, 0, 0, 0, 0, time.UTC)}
	err = order.Create()
	if err!=nil {
		panic(err)
	}
	err = order1.Create()
	if err!=nil {
		panic(err)
	}
	err = order2.Create()
	if err!=nil {
		panic(err)
	}
}

