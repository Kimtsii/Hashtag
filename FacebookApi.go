package main

import (
	"fmt"
	"os"

	// _ "os"
	"reflect"
	"strings"

	//"package main.go"

	fb "github.com/huandu/facebook/v2"
	"github.com/mehanizm/airtable"
)

var FbAccessToken string = os.Getenv("FACEBOOK_API")
var ATClientToken string = os.Getenv("AT_CLIENT_TOKEN")
var ATBaseID string = os.Getenv("AT_BASE_ID")

//var ATBaseIDa string = os.Getenv("CUSTOM")

// Tables
var ATFacebookPostsTable string = "Facebook _Post"
var ATHashtagsTable string = "Hashtags"

type FacebookPost struct {
	Id             string                 `facebook:",required"` // this field must exist in response.  // mind the "," before "required".
	FeedFrom       *FacebookPostFrom      `facebook:"from"`      // use customized field name "from".
	FeedFromShares *FacebookPostShares    `facebook:"shares"`
	CreatedTime    string                 `facebook:"created_time,required"` // both customized field name and "required" flag.
	Message        string                 `facebook:"message"`
	MsgHashTags    []string               `facebook:"-"` // Derived by parsing Message
	FeedFromReact  *FacebookPostReactions `facebook:"reactions"`
}

type FacebookPostFrom struct {
	Name string `json:"name"`                   // the "json" key also works as expected.
	Id   string `facebook:"id" json:"shadowed"` // if both "facebook" and "json" key are set, the "facebook" key is used.
}

type FacebookPostShares struct {
	Count int `facebook:"count"`
}

type FacebookPostReactions struct {
	Count    int    `facebook:"total_count"`
	Message1 string `facebook:"viewer_reaction"`
}

func parse_for_hashtags(Message string) []string {
	var hashtags []string
	// var hashtags = make([]string,0,100)

	possibletags := strings.Fields(Message) // Split message on spaces
	for _, value := range possibletags {
		if strings.HasPrefix(value, "#") {
			fmt.Println("found hashtag: ", value)
			hashtags = append(hashtags, value)
		}
	}

	// return strings.Fields(message)
	return hashtags
}

var AirTableClient *airtable.Client
var AirTableFbPostsTable *airtable.Table

var AirTableHashTagTable *airtable.Table

func init() {
	AirTableClient = airtable.NewClient(ATClientToken)
	AirTableFbPostsTable = AirTableClient.GetTable(ATBaseID, ATFacebookPostsTable)
	AirTableHashTagTable = AirTableClient.GetTable(ATBaseID, ATHashtagsTable)
}

func get_latest_fb_post() FacebookPost {
	var feed FacebookPost

	result, _ := fb.Get("107704657675864/feed?fields=reactions.summary(true),message,likes,created_time,shares,from&limit=2", fb.Params{
		"access_token": FbAccessToken,
	})
	result.DecodeField("data.0", &feed) // Unmarshal the JSON results into feed struct
	fmt.Println("latest post ID is: ", feed.Id)
	fmt.Println("latest post Created Time is: ", feed.CreatedTime)
	//fmt.Println("latest post share count is: ", feed.FeedFromShares.Count)
	fmt.Println("latest post message is: ", feed.Message)
	feed.MsgHashTags = parse_for_hashtags(feed.Message)
	//fmt.Println("latest post shares is", feed.FeedFromShares.Count)
	//fmt.Println("latest post react is", feed.FeedFromReact.Count)\

	return feed
}

// func main() {
// 	u := &User{
// 		Name: "Sammy",
// 	}

// 	fmt.Println(u)
// }

func main() {
	//

	{
		client := airtable.NewClient("keyOmJMHGYoQpMxYw")
		table := client.GetTable("appQntnFzrheCxlir", "Hashtags")

		records, err := table.GetRecords().
			FromView("Grid view").
			ReturnFields("Hashtag").
			Do()
		if err != nil {
			// Handle error
			panic(err)
		}

		for i := 0; i < len(records.Records); i++ {
			//	fmt.Print("Current Hashtags:", records.Records[i].Fields["Hashtag"], "\n")

		}

		feed := get_latest_fb_post()
		// initialize a slice literal
		//fmt.Println("Checking", feed.Id)
		fmt.Println("record type is: ", reflect.TypeOf(feed.Id))
		newSlice := records.Records
		hashtag := feed.MsgHashTags
		// fmt.Println("Data from Airtable:", newSlice)
		fmt.Println("Hashtags from post:", hashtag)
		if err == nil {
			fmt.Println("NO HASHTAG FROM POST")
		}

		for _, x := range hashtag {
			searchString := newSlice
			found := false
			fmt.Println("CHECKING HASHTAG ===>", x)
			for _, v := range searchString {
				//	fmt.Println("Hashtags:", v.Fields["Hashtag"])
				if x == v.Fields["Hashtag"] {
					found = true
					fmt.Println("HASHTAG", v.Fields["Hashtag"], "exist")
					fmt.Println("UPDATING HASHTAG...")

					m := v.Fields["Count"]
					//x++
					// strVar := "x"
					// intVar, err := strconv.Atoi(strVar)
					// fmt.Println(intVar, err, reflect.TypeOf(intVar))
					fmt.Println(m, "Check")
					fmt.Println("record type is: ", reflect.TypeOf(m))
					s := fmt.Sprintf("%f", m)
					fmt.Println("record type is: ", reflect.TypeOf(s), s)

					strVar := s
					intValue := 0
					_, err := fmt.Sscan(strVar, &intValue)
					fmt.Println(intValue, err, reflect.TypeOf(intValue))
					//fmt.Println("record type is: ", reflect.TypeOf(s), s)
					intValue++

					toUpdateRecords := &airtable.Records{
						Records: []*airtable.Record{

							{
								ID: v.ID,
								Fields: map[string]interface{}{

									"Hashtag":   v.Fields["Hashtag"],
									"Last Used": feed.CreatedTime,
									"Count":     intValue,
									"Feed_ID":   feed.Id,
								},
							},
						},
					}
					updatedRecords, err := table.UpdateRecords(toUpdateRecords)
					if err != nil {
						// Handle error
						panic(err)
					}

					for i := 0; i < len(toUpdateRecords.Records); i++ {
						fmt.Println(updatedRecords.Records[i].ID)
					}

				}

			}
			if !found {
				fmt.Println("HASHTAG", x, " does NOT exist")
				fmt.Println("ADDING IT NOW")
				//feed := get_latest_fb_post()

				recordsToSend := &airtable.Records{
					Records: []*airtable.Record{
						{
							Fields: map[string]interface{}{
								"Hashtag":   x,
								"Last Used": feed.CreatedTime,
								"Count":     1,
								"Feed_ID":   feed.Id,
								//"Record ID":       v.ID,
								//"Created Time": feed.CreatedTime,
							},
						},
					},
				}

				receivedRecords, err := table.AddRecords(recordsToSend)
				fmt.Println(recordsToSend)
				//fmt.Println(reflect.TypeOf(err))
				//fmt.Println(reflect.TypeOf(receivedRecords))
				if err != nil {
					fmt.Println("Error writing records: ", err)
				}
				fmt.Println(receivedRecords)
			}
		}

	}

}
