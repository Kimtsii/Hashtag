package main

import (
	"fmt"
	//"reflect"
	"strings"

	fb "github.com/huandu/facebook/v2"
	"github.com/mehanizm/airtable"
)

// You have to set your environment variables properly for either MacOS, Linux, or Windows
var FbAccessToken string = ("EAAHPmz80hhABANKttQHhzco4QaPGezGVWbSkmUcQVQZC1ru7doNqOkhu402ztuEhOJHw1TjlQNf5G8uiqp6cGN3z3AlpVM3eQmzOsvUqUXTHBRFoMfMDXD7LqfVD9ShUMmfxVyiZBjLwiau3kqJ6mCb9vL8gh8mELQyCrURAzxsBjsILNGwG1yHfCaEf8FVxxMSkQcSS89fZCpWzIP4BWXfwiAHn8E4YSZC9iMioUGFr0kgxYII9	")
var ATClientToken string = ("keyOmJMHGYoQpMxYw")
var ATBaseID string = ("appQntnFzrheCxlir")

// Tables
var ATFacebookPostsTable string = "Testing"
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
	Count   int    `facebook:"total_count"`
	Message string `facebook:"viewer_reaction"`
}

func parse_for_hashtags(message string) []string {
	var hashtags []string
	// var hashtags = make([]string,0,100)

	possibletags := strings.Fields(message) // Split message on spaces
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

	result, _ := fb.Get("me/feed?limit=1", fb.Params{
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

// func check_for_fb_post_in_airtable(facebook_id string) (bool, *airtable.Record) {

// 	record := new(airtable.Record)
// 	fmt.Println("checking if the hashtag: ", facebook_id, " exist")
// 	records, err := AirTableHashTagTable.GetRecords().
// 		FromView("Grid view").
// 		ReturnFields("Hashtag").
// 		MaxRecords(1).
// 		Do()
// 	fmt.Println("records length is ", len(records.Records))
// 	//fmt.Println("err is ", err)

// 	if err == nil && len(records.Records) == 1 {
// 		record = records.Records[0]
// 		fmt.Println("record type is: ", reflect.TypeOf(record))
// 		fmt.Println("record is: ", record)
// 		return true, record
// 	} else {
// 		fmt.Println("received error: ", err)
// 		return false, record
// 	}
// }

// func contains(MsgHashTags []string, v string) bool {
// 	for _, s := range MsgHashTags {
// 		if v == s {
// 			return true
// 		}
// 	}
// 	return false
// }

//  Really need to test for a whole bunch of different conditions

func main() {

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
			fmt.Print(records.Records[i].Fields, "\n")

		}

		feed := get_latest_fb_post()
		var hashtagCount = len(feed.MsgHashTags)
		fmt.Println(feed.CreatedTime)

		fmt.Println("no existing record, adding it")

		//	for i := 0; i < len(records.Records); i++
		//for i, element1 := range records.Records

		//	for i, element := range records.Records
		for i := 0; i < hashtagCount; i++ {
			{

				var arr = records.Records[i].Fields
				//	fmt.Println("RECORD:", records.Records[i].Fields)
				var element1 string = feed.MsgHashTags[2]
				//	fmt.Println("MESSAGES:", feed.MsgHashTags[i])

				var result bool = false
				for _, x := range arr {
					if x == element1 {
						result = true
						break
					}
				}

				if result {
					fmt.Println("Element", element1, "is present in the array:", arr)
					fmt.Println("Checking if its newer or not")

					if feed.CreatedTime == records.Records[i].CreatedTime {
						fmt.Println("TRUE")
					} else {
						fmt.Println("VALUE IS NOT NEW, UPDATING THE DATE...")
					}

					// toUpdateRecords := &airtable.Records{
					// 	Records: []*airtable.Record{

					// 		{
					// 			ID: records.Records[i].ID,
					// 			Fields: map[string]interface{}{
					// 				//"FacebookID":   feed.Id,
					// 				// "Message":      feed.Message,
					// 				// "Created Time": feed.CreatedTime,
					// 				"Count": "FOUND",
					// 				//"Created Time": feed.CreatedTime,
					// 			},
					// 		},
					// 	},
					// }
					// updatedRecords, err := AirTableFbPostsTable.UpdateRecords(toUpdateRecords)
					// if err != nil {
					// 	// Handle error
					// 	panic(err)
					// }

					// for i := 0; i < len(toUpdateRecords.Records); i++ {
					// 	fmt.Print(updatedRecords.Records[i].ID)
					// }
					break

				} else {
					fmt.Println("Element", element1, "is not present in the array:", arr)
					//	fmt.Println(records.Records[i].Fields)

					// recordsToSend := &airtable.Records{
					// 	Records: []*airtable.Record{
					// 		{
					// 			Fields: map[string]interface{}{
					// 				//"Hashtag": feed.MsgHashTags,
					// 				"Hashtag": feed.MsgHashTags[i],
					// 				//"Shares":       feed.FeedFromShares.Count,
					// 				"Last Used": feed.CreatedTime,
					// 			},
					// 		},
					// 	},
					// }

					// receivedRecords, err := AirTableHashTagTable.AddRecords(recordsToSend)
					// //fmt.Println(recordsToSend)
					// //fmt.Println(reflect.TypeOf(err), "No error")
					// //fmt.Println(reflect.TypeOf(receivedRecords), "check")
					// if err != nil {
					// 	fmt.Println("Error writing records: ", err)
					// }

					// for i := 0; i < len(receivedRecords.Records); i++ {
					// 	//	fmt.Print(receivedRecords.Records[i].Fields["Hashtag"], "\n")

					// }
				}
				// if feed.MsgHashTags[i] == records.Records[i].Fields["Hashtag"] {
				// 	fmt.Println("FOUND IT!")
				// } else {
				// 	fmt.Println("Happy", records.Records[i].Fields["Hashtag"])
				// }

				//	fmt.Println(contains([]string{feed.MsgHashTags[i]}, feed.MsgHashTags[i]))
				//	fmt.Println(feed.MsgHashTags[i])

			}
		}

		//for reference
		// arr := [6]string{"haha", "sip", "locked"}
		// var element string = "hah2a"

		// var result bool = false
		// for _, x := range arr {
		// 	if x == element {
		// 		result = true
		// 		break
		// 	}
		// }

		// if result {
		// 	fmt.Print("Element is present in the array.")
		// } else {
		// 	fmt.Print("Element is not present in the array.")
		// }
		// 	{
		// 	client := airtable.NewClient("keyOmJMHGYoQpMxYw")
		// 	table := client.GetTable("appQntnFzrheCxlir", "Hashtags")

		// 	records, err := table.GetRecords().
		// 		FromView("Grid view").
		// 		ReturnFields("Hashtag").
		// 		Do()
		// 	if err != nil {
		// 		// Handle error
		// 		panic(err)
		// 	} else {
		// 		for i := 0; i < hashtagCount; i++ {
		// 			if feed.MsgHashTags[i] == records.Records[i].Fields["Hashtag"] {
		// 				fmt.Println("Hashtag that exist: ", feed.MsgHashTags[i])

		// 			} else {
		// 				toUpdateRecords := &airtable.Records{
		// 					Records: []*airtable.Record{
		// 						{
		// 							ID: existing_fb_post.ID,
		// 							Fields: map[string]interface{}{
		// 								"Hashtag": feed.MsgHashTags[i],
		// 								//"Message":      feed.Message,
		// 								"Last Used": feed.CreatedTime,
		// 								//"Shares":       feed.FeedFromShares.Count
		// 							},
		// 						},
		// 					},
		// 				}
		// 				updatedRecords, err := AirTableHashTagTable.UpdateRecords(toUpdateRecords)
		// 				if err != nil {
		// 					// Handle error

		// 					panic(err)

		// 				}
		// 				for i := 0; i < len(toUpdateRecords.Records); i++ {
		// 					fmt.Println(updatedRecords.Records[i].Fields["Hashtag"])
		// 				}

		// 			}

		// 		}
		// 	}

		// 	// toUpdateRecords := &airtable.Records{
		// 	// 	Records: []*airtable.Record{

		// 	// 		{
		// 	// 			ID: existing_fb_post.ID,
		// 	// 			Fields: map[string]interface{}{
		// 	// 				"Notes": feed.MsgHashTags[0],
		// 	// 				//"Message":      feed.Message,
		// 	// 				"Last Used": feed.CreatedTime,
		// 	// 				//"Shares":       feed.FeedFromShares.Count

		// 	// 			},
		// 	// 		},
		// 	// 	},
		// 	// }
		// 	// updatedRecords, err := AirTableHashTagTable.UpdateRecords(toUpdateRecords)
		// 	// if err != nil {
		// 	// 	// Handle error
		// 	// 	panic(err)
		// 	// }

		// 	// for i := 0; i < len(toUpdateRecords.Records); i++ {
		// 	// 	fmt.Print(updatedRecords.Records[i].ID)
		// 	// }

		// }

	}

	// fmt.Println(AirTableHashTagTable.GetRecords().
	// 	FromView("Grid view").
	// 	ReturnFields("Hashtag").
	// 	Do())

}
