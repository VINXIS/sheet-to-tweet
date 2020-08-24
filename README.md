# sheet-to-tweet
A bot that posts from a spreadsheet to twitter

1. Obtain the credentials json from Google Cloud console by first enabling sheets API, and then creating a service account. Place it in the config folder and rename it to credentials.json.

2. Create a twitter app in the twitter developer page and obtain its consumer key, consumer secret, api key, and api secret.

3. Your spreadsheet should have 2 columns, column A with the text that will be tweeted, and column B that states if it is posted or not. The 2nd column should be empty for ones that have not been posted yet. Make the first row have the headers, something like Tweet for column A and Posted? for column B.

4. Create a config.json using the example format of config.example.json, and then input the sheet and range of what you want to tweet in as well. It will assume the text to tweet is in the first column.

5. Run the bot.