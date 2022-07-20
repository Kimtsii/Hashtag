package main

import (
	"fmt"

	//"reflect"
	"strings"

	fb "github.com/huandu/facebook/v2"
	"github.com/mehanizm/airtable"
)

// You have to set your environment variables properly for either MacOS, Linux, or Windows
var FbAccessToken string = ("EAAHPmz80hhABALkKpAoZC2wzxX6KPW1THowS4QYjw823Fq012Bym1zZA1OZCu4N7oIsxcEUvebLFeaCiwqpESBCZATHqepMorRSXVZCzRmHbtrMg1JYV6Iq8vhnlKKNYw21tCYPnZCkKFh7t4GZA6YRayfZCRzLl7ETsK59vdFNNTA5natciaDLJw5H2McyuRLn1cBc7pX09SajsZCjrPaQGZCZAIjHPPXhOeMQV68aFLc0skoyzBcnJC1a		")
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
			//fmt.Println("found hashtag: ", value)
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

func contains(s []string, v string) bool {
	for _, s := range s {
		if v == s {
			return true
		}
	}
	return false
}

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
			fmt.Print("Current Hashtags:", records.Records[i].Fields["Hashtag"], "\n")

		}

		feed := get_latest_fb_post()
		// initialize a slice literal
		newSlice := records.Records
		hashtag := feed.MsgHashTags
		// fmt.Println("Data from Airtable:", newSlice)
		fmt.Println("Data from post:", hashtag)

		// initialize the strings to search for

		// initialize a found flag

		//searchString := hashtag
		for _, x := range newSlice {
			searchString := hashtag
			found := false
			fmt.Println("CHECKING HASHTAG ===>", x.Fields["Hashtag"])
			for _, v := range searchString {
				fmt.Println("Hashtags:", v)
				if v == x.Fields["Hashtag"] {
					found = true
					fmt.Println("HASHTAG", x.Fields["Hashtag"], "exist")
					fmt.Println("UPDATING HASHTAG...")

					toUpdateRecords := &airtable.Records{
						Records: []*airtable.Record{

							{
								ID: x.ID,
								Fields: map[string]interface{}{

									"Hashtag":   x.Fields["Hashtag"],
									"Last Used": feed.CreatedTime,
									"Count":     "xaxa",
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
				// check if the strings match
				//    if v == feed.Msghashtags  {
				// 	 found = true
				// 	 fmt.Println("The slice contains", searchString, "at index", i)
				// 	 break
				//    }
			}
			if found == false {
				fmt.Println("HASHTAG", x.Fields["Hashtag"], " does NOT exist")
			}

		}
		// if string not found

		//feed := get_latest_fb_post()
		//var hashtagCount = len(feed.MsgHashTags)

		//	for i := 0; i < len(records.Records); i++
		//for i, element1 := range records.Records

		//	for i, element := range records.Records

		// for _, element1 := range feed.MsgHashTags {
		// 	fmt.Println("Checking if hashtag:", element1, "is present")
		// 	//fmt.Println("Type of", reflect.TypeOf(feed.MsgHashTags))

		// 	for _, element := range records.Records {

		// 		// fmt.Println(element.Fields["Hashtag"])
		// 		fmt.Println("Is hashtag:", element.Fields["Hashtag"], "equal to", element1)

		// 		if element1 == element.Fields["Hashtag"] {
		// 			fmt.Println("YES")
		// 			fmt.Println("RECORD FOUND UPDATING IT:", element)
		// 			fmt.Println(element.ID)

		// 			toUpdateRecords := &airtable.Records{
		// 				Records: []*airtable.Record{

		// 					{
		// 						ID: element.ID,
		// 						Fields: map[string]interface{}{

		// 							"Hashtag":   element.Fields["Hashtag"],
		// 							"Last Used": feed.CreatedTime,
		// 						},
		// 					},
		// 				},
		// 			}
		// 			updatedRecords, err := table.UpdateRecords(toUpdateRecords)
		// 			if err != nil {
		// 				// Handle error
		// 				panic(err)
		// 			}

		// 			for i := 0; i < len(toUpdateRecords.Records); i++ {
		// 				fmt.Println(updatedRecords.Records[i].ID)
		// 			}

		// 		} else {

		// 			fmt.Println("NO")

		// 		}

		// 	}

		// }
		// for _, element := range records.Records {
		// 	fmt.Println(contains(element, feed.MsgHashTags[1]))
		// 	//fmt.Println(contains(s, "#Jack"))
		// }

		// var result bool = false
		// for _, x := range records.Records {
		// 	if x == e {
		// 		result = true
		// 		break
		// 	}
		// 	if result {
		// 		fmt.Println("PRINT", x)
		// 	} else {
		// 		fmt.Println("PRINT ALL THE WAY")
		// 	}

		// // }
		// }

		// How to distinguish between the two?

		// m1 := map[string]string{"a": feed.MsgHashTags[i]}
		// m2 := map[*airtable.Record]string{"a": records.Records[i].Fields}
		// // "a" exists in map
		// v1 := m1["a"]
		// fmt.Println("HOOOOOOOy", v1) // 0

		// // "f" does not exist in the map
		// v2 := m2["a"]
		// fmt.Println("Check", v2) // 0

		// var arr = records.Records[i].Fields
		// //fmt.Println("RECORD:", records.Records[i].Fields)

		// var element1 string = feed.MsgHashTags[i]
		// //fmt.Println("MESSAGES:", feed.MsgHashTags[i])

		// var result bool = false
		// for _, x := range arr {
		// 	if x == element1 {
		// 		result = true
		// 		break
		// 	}
		// }

		// if result {
		// 	fmt.Println("Element", element1, "is present in the array:", arr)
		// 	fmt.Println("Checking if its newer or not")

		// 	if feed.CreatedTime == records.Records[i].CreatedTime {
		// 		fmt.Println("VALUE IS NEW... DO NOTHING")
		// 	} else {

		// 		fmt.Println("VALUE IS NOT NEW, UPDATING THE DATE...")
		// 		for x, element := range arr {
		// 			fmt.Println("At index", x, "value is", element)
		// 		}

		// 		// toUpdateRecords := &airtable.Records{
		// 		// 	Records: []*airtable.Record{

		// 		// 		{
		// 		// 			ID: records.Records[i].ID,
		// 		// 			Fields: map[string]interface{}{
		// 		// 				//"FacebookID":   feed.Id,
		// 		// 				// "Message":      feed.Message,
		// 		// 				// "Created Time": feed.CreatedTime,
		// 		// 				"Count": "FOUND",
		// 		// 				//"Created Time": feed.CreatedTime,
		// 		// 			},
		// 		// 		},
		// 		// 	},
		// 		// }
		// 		// updatedRecords, err := AirTableFbPostsTable.UpdateRecords(toUpdateRecords)
		// 		// if err != nil {
		// 		// 	// Handle error
		// 		// 	panic(err)
		// 		// }

		// 		// for i := 0; i < len(toUpdateRecords.Records); i++ {
		// 		// 	fmt.Print(updatedRecords.Records[i].ID)
		// 		// }
		// 	}

		// 	// break

		// } else {
		// 	fmt.Println("Element", element1, "is not present in the array:", arr)
		// 	fmt.Println("ADDING IT NOW!")
		// 	//	fmt.Println(records.Records[i].Fields)

		// 	// recordsToSend := &airtable.Records{
		// 	// 	Records: []*airtable.Record{
		// 	// 		{
		// 	// 			Fields: map[string]interface{}{
		// 	// 				//"Hashtag": feed.MsgHashTags,
		// 	// 				"Hashtag": feed.MsgHashTags[i],
		// 	// 				//"Shares":       feed.FeedFromShares.Count,
		// 	// 				"Last Used": feed.CreatedTime,
		// 	// 			},
		// 	// 		},
		// 	// 	},
		// 	// }

		// 	// receivedRecords, err := AirTableHashTagTable.AddRecords(recordsToSend)
		// 	// //fmt.Println(recordsToSend)
		// 	// //fmt.Println(reflect.TypeOf(err), "No error")
		// 	// //fmt.Println(reflect.TypeOf(receivedRecords), "check")
		// 	// if err != nil {
		// 	// 	fmt.Println("Error writing records: ", err)
		// 	// }

		// 	// for i := 0; i < len(receivedRecords.Records); i++ {
		// 	// 	//	fmt.Print(receivedRecords.Records[i].Fields["Hashtag"], "\n")

		// 	// }
		// }

		//}
		//	}

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
