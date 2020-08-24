package main

import (
	"context"
	"encoding/base64"
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
)

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
	conf := config.NewConfig()
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

	ticker := time.NewTicker(timer)
	log.Printf("Beginning timer! First tweet will be posted %v from now.\n", timer)
	for {
		select {
		case <-ticker.C:
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
					if rowNum != -1 {
						category := fmt.Sprintf("%v", row[2])
						if category == "Current Events" {
							continue
						} else {
							target = row
							rowNum = i + 1
						}
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
			t, err := twitter.PostTweet(text, v)
			if err != nil {
				log.Fatalf("Unable to post tweet: %v", err)
			} else {
				log.Printf("Posted tweet ID: %v\n", t.Id)
				cell := "Approved!D" + strconv.Itoa(rowNum)
				_, err = sheetsService.Spreadsheets.Values.Update(conf.Sheet.ID, cell, &sheets.ValueRange{Values: [][]interface{}{{"Y"}}}).ValueInputOption("RAW").Do()
				if err != nil {
					log.Fatalf("Failed to update sheet: %v", err)
				}
			}
		}
	}
}
