feedme is a small command-line feed reader.

*feedme* aims to be fast and simple. Run it and it spits out a list of new posts (title+url) since the last run.

- `feedme add [URL]`: add one or more feeds to the fetch list
- `feedme delete [URL]`: delete one or more feeds from the fetch list
- `feedme list`: list all the fetched feeds
- `feedme fetch` for each feed, print the title+URL of each new item

# Possible improvements

- replace the feedparser dependency with custom code based on encoding/xml, in order to parse only the new posts
- email the results, like rss2email ? this could be done by piping the output to a MDA ...

