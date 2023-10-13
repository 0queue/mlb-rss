# mlb-rss

A program to generate an RSS feed from `statsapi.mlb.com`

See someone else's swagger.json [here](https://github.com/joerex1418/mlb-statsapi-swagger-docs)

## games.json

An example API response from

`https://statsapi.mlb.com/api/v1/schedule/games/?sportId=1&startDate=2022-04-15&endDate=2022-04-22`

> TODO add hydrate=game(content(highlights(all))) to query (see swagger)

## Building

```
just build
```

## Publish

```
just ko 
```

## Usage


Service runs by default on port 8080 serving the /rss.xml path. When using the container image,
make sure to set `JSON_LOG=true`

```
just r
```
  
## My deployment

Basically a hello world nomad job