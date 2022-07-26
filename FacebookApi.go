package main

import (
	"fmt"

	//"reflect"
	"strings"

	fb "github.com/huandu/facebook/v2"
	"github.com/mehanizm/airtable"
)

// You have to set your environment variables properly for either MacOS, Linux, or Windows
var FbAccessToken string = ("EAAHPmz80hhABALeWt0zTDFOVXNUEPG9noL6DSRC40YxnhDSqK4rTFJ8TTYHS3dTjfXCer1inX7PkaqZAFo8XoZAaexcHQd2YC68xdcCAz5BXzWRh9O6yp9Il97TWwIIiYiRZAGaBrCLCtEJVeC9cTRIwN1z90NOmwjbsloBSZAGTK8Pf8WRD		")
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
		fmt.Println("Hashtags from post:", hashtag)
		if err == nil {
			fmt.Println("NO HASHTAG FROM POST")
		}

		for _, x := range hashtag {
			searchString := newSlice
			found := false
			fmt.Println("CHECKING HASHTAG ===>", x)
			for _, v := range searchString {
				fmt.Println("Hashtags:", v.Fields["Hashtag"])
				if x == v.Fields["Hashtag"] {
					found = true
					fmt.Println("HASHTAG", v.Fields["Hashtag"], "exist")
					fmt.Println("UPDATING HASHTAG...")

					toUpdateRecords := &airtable.Records{
						Records: []*airtable.Record{

							{
								ID: v.ID,
								Fields: map[string]interface{}{

									"Hashtag":   v.Fields["Hashtag"],
									"Last Used": feed.CreatedTime,
									"Count":     "UPDATED",
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
			if found == false {
				fmt.Println("HASHTAG", x, " does NOT exist")
				fmt.Println("ADDING IT NOW")
				//feed := get_latest_fb_post()

				recordsToSend := &airtable.Records{
					Records: []*airtable.Record{
						{
							Fields: map[string]interface{}{
								"Hashtag":   x,
								"Last Used": feed.CreatedTime,
								"Count":     "ADDED",
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
