package config

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/sheets/v4"
)

// Config holds the main configuration information
type Config struct {
	Sheet   Sheet
	Twitter Twitter
	Google  jwt.Config
}

// Twitter holds the information for the twitter API
type Twitter struct {
	Token          string
	Secret         string
	ConsumerKey    string
	ConsumerSecret string
}

// Sheet holds the information for the spreadsheed
type Sheet struct {
	ID    string
	Range string
}

// NewConfig creates a new config struct for you based off of the json file
func NewConfig() *Config {
	b, err := ioutil.ReadFile("./config/config.json")
	if err != nil {
		log.Fatal("An error occured in obtaining configuration information: ", err)
	}
	config := &Config{}
	json.Unmarshal(b, config)

	// Obtain google config
	b, err = ioutil.ReadFile("./config/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	gConf, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to create JWTConfig file: %v", err)
	}
	config.Google = *gConf

	return config
}
