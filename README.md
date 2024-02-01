# SearchHouse Spider
Web crawlers are dime-a-dozen. I built this spider specifically to retrieve web pages
at a high volume as the beginning of my search engine pipeline.

## Features
- Spawn an arbitrary amount of routines to crawl pages
  - Each routine has its own delay to honor politeness
- Store frontier in SQLite database
  - Safe for concurrency
  - Crawler resumes where it left off if closed
  - Since the biggest bottleneck is IO (networking), retrieving frontier from disk is still viable
- Scheduler in main execution will allocate domains to each routine based on hash
  - Similar to a hashmap data structure
- Store pages, server responses, and timestamp as JSON object
- Hash URLs to store pages efficiently on disk

## Installation
```
git clone https://github.com/marceloclubhouse/searchhouse-spider
cd searchhouse-spider
go install searchhouse-spider
go run main.go -h
```

## Running
```
go run main.go -seed="https://marcelocubillos.com" -numRoutines=100
```

## Flow
After specifying the values for the crawler to use, just execute main.go and
watch the spider take off!