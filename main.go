package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"./config"
	"github.com/ChimeraCoder/anaconda"

	"google.golang.org/api/sheets/v4"
)

var (
	linkRegex  = regexp.MustCompile(`(?i)https?:\/\/\S+`)
	rangeRegex = regexp.MustCompile(`(.+)![A-Z][[:digit:]]*:[A-Z][[:digit:]]*`)
	conf       = config.NewConfig()
)

type twitterError struct {
	Errors []errorMessage `json:"errors"`
}

type errorMessage struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	// Get args
	timer := 30 * time.Minute
	for i, arg := range os.Args {
		if arg == "-t" {
			num, err := strconv.Atoi(os.Args[i+1])
			if err != nil {
				log.Fatalln("Please provide an integer with the -t flag.")
			}
			timer = time.Duration(num) * time.Minute
		}
	}
	log.Printf("Time interval obtained: %v\n", timer)

	// Get sheets/twitter credentials + sheet ID/range
	if !rangeRegex.MatchString(conf.Sheet.Range) {
		log.Fatalln("Range given in config is not a valid range!")
	}
	log.Println("Obtained config info...")

	// Enable google client
	client := conf.Google.Client(context.TODO())
	sheetsService, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to create Sheet Service: %v", err)
	}
	log.Println("Started google client...")

	// Enable twitter client
	twitter := anaconda.NewTwitterApiWithCredentials(conf.Twitter.Token, conf.Twitter.Secret, conf.Twitter.ConsumerKey, conf.Twitter.ConsumerSecret)
	log.Println("Started twitter client...")

	currTime := time.Now()
	nextHour := time.Date(currTime.Year(), currTime.Month(), currTime.Day(), currTime.Hour()+1, 0, 0, 0, time.UTC)

	delay := time.NewTimer(nextHour.Sub(time.Now()))
	log.Printf("Waiting for the next hour to pass... First tweet will be posted on %v", nextHour)
	<-delay.C
	go runCron(sheetsService, twitter)

	// Run cron
	ticker := time.NewTicker(timer)
	log.Printf("Beginning timer! Next tweet will be posted %v from now.\n", timer)
	for {
		select {
		case <-ticker.C:
			go runCron(sheetsService, twitter)
		}
	}
}

func runCron(sheetsService *sheets.Service, twitter *anaconda.TwitterApi) {
	// Obtain sheet values
	res, err := sheetsService.Spreadsheets.Values.Get(conf.Sheet.ID, conf.Sheet.Range).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}
	colCount := len(res.Values[0])
	rowNum := -1
	target := []interface{}{}
	for i, row := range res.Values {
		if len(row) != colCount {
			if rowNum != -1 { // In case there are current events to post first as they take higher priority
				category := fmt.Sprintf("%v", row[2])
				if category != "Current Events" {
					continue
				}

				target = row
				rowNum = i + 1
				break
			}
			target = row
			rowNum = i + 1
		}
	}
	if rowNum == -1 {
		log.Fatalln("No more new rows to send.")
	}
	text := fmt.Sprintf("%v", target[1])
	v := url.Values{}

	// Check for a link for an image or video
	if linkRegex.MatchString(text) {
		link := linkRegex.FindStringSubmatch(text)[0]
		res, err := http.Get(link)
		if err == nil {
			defer res.Body.Close()
			b, err := ioutil.ReadAll(res.Body)
			if err == nil {
				media, err := twitter.UploadMedia(base64.StdEncoding.EncodeToString(b))
				if err == nil {
					v.Set("media_ids", strconv.FormatInt(media.MediaID, 10))
					text = strings.Replace(text, link, "", -1)
				}
			}
		}
	}

	// Post tweet
	t, err := twitter.PostTweet(text, v)
	if err != nil {
		var twitError twitterError
		ogErr := err
		err = json.Unmarshal([]byte(err.Error()), &twitError)
		if err != nil {
			log.Fatalf("Unable to parse error JSON: %v\n Original error: %v", err, ogErr)
		}
		if twitError.Errors[0].Code == 187 { // Duplicate error code
			log.Println("Tweet was duplicate, finding next tweet...")
			updateSheet(sheetsService, rowNum)
			go runCron(sheetsService, twitter)
		} else {
			log.Fatalf("Error in posting tweet: %v", err)
		}
	} else {
		log.Printf("Posted tweet ID: %v\n", t.Id)
		updateSheet(sheetsService, rowNum)
	}
}

// Y the row that has been posted
func updateSheet(sheetsService *sheets.Service, rowNum int) {
	cell := "Approved!D" + strconv.Itoa(rowNum)
	_, err := sheetsService.Spreadsheets.Values.Update(conf.Sheet.ID, cell, &sheets.ValueRange{Values: [][]interface{}{{"Y"}}}).ValueInputOption("RAW").Do()
	if err != nil {
		log.Fatalf("Failed to update sheet: %v", err)
	}
}
