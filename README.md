# sheet-to-tweet
A bot that posts from a spreadsheet to twitter. this branch is for osu! anonymous (https://twitter.com/anonymous_osu)

1. Obtain the credentials json from Google Cloud console by first enabling sheets API, and then creating a service account. Place it in the config folder and rename it to credentials.json.

2. Create a twitter app in the twitter developer page and obtain its consumer key, consumer secret, api key, and api secret.

3. Your spreadsheet should have 4 columns, column B with the text that will be tweeted, and column D that states if it is posted or not. The first is for date submitted and the third is submission category type. For any submission that should be pushed to the front of the queue, place `Current Events` for the column C cell for that row. Last column should be kept empty, the bot will write Y for any tweet posted.

4. Create a config.json using the example format of config.example.json, and then input the sheet and range of what you want to tweet in as well. It will assume the text to tweet is in the first column.

5. Run the bot.
