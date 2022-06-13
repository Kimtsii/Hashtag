package main

import (
	"fmt"
	"reflect"
	"strings"

	fb "github.com/huandu/facebook/v2"
	"github.com/mehanizm/airtable"
)

// You have to set your environment variables properly for either MacOS, Linux, or Windows
var FbAccessToken string = ("EAAHPmz80hhABALkX9tuCUiV289VCN1sAhDUk83Fbw3mEgNCUj6I4LFix1hARjDxy1XNZBcAgMeL5nJ4ZC2JHckKXOZBmysFdw5ZAA98voApVYkZAiXc8nBVX8RuPWeh6t8x3tz8GAk3fr1Vcnv41GjKeuCY22pBQHEo62OR01g8XdrlZBfCcjSGWSBsnO5mXy54kkNPvnhDwpn8ZConncTtRdjtrMuvMAcFSvsf263nVLkQueoEAaY1")
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
	fmt.Println("ssss")
	for _, value := range possibletags {
		fmt.Println(value)
		if strings.HasPrefix(value, "#") {
			fmt.Println("found hashtag: ", value)
			hashtags = append(hashtags, value)

		}
	}

	fmt.Println(strings.Fields(message))
	fmt.Println("test: ", hashtags)

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
	//fmt.Println("latest post react is", feed.FeedFromReact.Count)
	fmt.Println(feed.MsgHashTags)
	fmt.Println("wtf")

	return feed
}

func check_for_fb_post_in_airtable(facebook_id string) (bool, *airtable.Record) {
	record := new(airtable.Record)
	fmt.Println("checking if the hashtag: ", facebook_id, " already exists")
	records, err := AirTableHashTagTable.GetRecords().
		FromView("Grid view").
		ReturnFields("Hashtag").
		MaxRecords(1).
		Do()
	fmt.Println("records length is ", len(records.Records))
	fmt.Println("err is ", err)

	if err == nil && len(records.Records) == 1 {
		record = records.Records[0]
		fmt.Println("record type is: ", reflect.TypeOf(record))
		fmt.Println("record is: ", record)
		return true, record
	} else {
		fmt.Println("received error: ", err)
		return false, record
	}
}

//  Really need to test for a whole bunch of different conditions

func main() {

	{
		feed := get_latest_fb_post()
		exists, existing_fb_post := check_for_fb_post_in_airtable(feed.Id)
		fmt.Println("Hashtag exists is ", exists)

		fmt.Println("www ")
		fmt.Println(feed.MsgHashTags[0])
		var hashtagCount = len(feed.MsgHashTags)
		fmt.Println(hashtagCount)

		if exists == false {
			fmt.Println("no existing record, adding it")

			for i := 0; i < hashtagCount; i++ {
				fmt.Println("Hastag: ", feed.MsgHashTags[i])
				recordsToSend := &airtable.Records{
					Records: []*airtable.Record{
						{
							Fields: map[string]interface{}{
								//"Hashtag": feed.MsgHashTags,
								"Notes": feed.MsgHashTags[i],
								//"Shares":       feed.FeedFromShares.Count,
								"Last Used": feed.CreatedTime,
							},
						},
					},
				}

				receivedRecords, err := AirTableHashTagTable.AddRecords(recordsToSend)
				fmt.Println(recordsToSend)
				fmt.Println(reflect.TypeOf(err))
				fmt.Println(reflect.TypeOf(receivedRecords))
				if err != nil {
					fmt.Println("Error writing records: ", err)
				}
				fmt.Println(receivedRecords)
			}

			// recordsToSend := &airtable.Records{
			// 	Records: []*airtable.Record{
			// 		{
			// 			Fields: map[string]interface{}{
			// 				//"Hashtag": feed.MsgHashTags,
			// 				"Notes": feed.MsgHashTags[0],
			// 				//"Shares":       feed.FeedFromShares.Count,
			// 				"Last Used": feed.CreatedTime,
			// 			},
			// 		},
			// 	},
			// }

			// receivedRecords, err := AirTableHashTagTable.AddRecords(recordsToSend)
			// fmt.Println(recordsToSend)
			// fmt.Println(reflect.TypeOf(err))
			// fmt.Println(reflect.TypeOf(receivedRecords))
			// if err != nil {
			// 	fmt.Println("Error writing records: ", err)
			// }
			// fmt.Println(receivedRecords)
		} else {
			fmt.Println("record found updating it", existing_fb_post)
			fmt.Println("record found, " + existing_fb_post.ID + " updating it")

			for i := 0; i < hashtagCount; i++ {
				fmt.Println("Hastag: ", feed.MsgHashTags[i])
				toUpdateRecords := &airtable.Records{
					Records: []*airtable.Record{

						{
							ID: existing_fb_post.ID,
							Fields: map[string]interface{}{
								"Notes": feed.MsgHashTags[i],
								//"Message":      feed.Message,
								"Last Used": feed.CreatedTime,
								//"Shares":       feed.FeedFromShares.Count

							},
						},
					},
				}
				updatedRecords, err := AirTableHashTagTable.UpdateRecords(toUpdateRecords)
				if err != nil {
					// Handle error
					panic(err)
				}

				for i := 0; i < len(toUpdateRecords.Records); i++ {
					fmt.Print(updatedRecords.Records[i].ID)
				}
			}

			// toUpdateRecords := &airtable.Records{
			// 	Records: []*airtable.Record{

			// 		{
			// 			ID: existing_fb_post.ID,
			// 			Fields: map[string]interface{}{
			// 				"Notes": feed.MsgHashTags[0],
			// 				//"Message":      feed.Message,
			// 				"Last Used": feed.CreatedTime,
			// 				//"Shares":       feed.FeedFromShares.Count

			// 			},
			// 		},
			// 	},
			// }
			// updatedRecords, err := AirTableHashTagTable.UpdateRecords(toUpdateRecords)
			// if err != nil {
			// 	// Handle error
			// 	panic(err)
			// }

			// for i := 0; i < len(toUpdateRecords.Records); i++ {
			// 	fmt.Print(updatedRecords.Records[i].ID)
			// }

		}

	}

}
