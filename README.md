# kinohub-core

## API

### Search media content

`GET /search?q=`

### Get detailed information about media item

`GET /items/:item-id`

### Get TV Shows releases

`GET /tv/releases?from=2017-08-15&to=2017-08-24`

### Get TV Shows

`GET /tv/watching`

## Development

## Downloader (?)

Requiremets:

- predownload files to play from local FS
- automatically cleanup old file
- ability to set max size of stored data

### Overview

Downloader - tie metadata / source and local file
DownloadQueue - queue of meta/sources to download
FileStorage - to allocate new and cleanup old files
Streamer - stream file content
