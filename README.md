# mlb-rss

A program to generate an RSS feed from `statsapi.mlb.com`

## games.json

An example API response from

`https://statsapi.mlb.com/api/v1/schedule/games/?sportId=1&startDate=2022-04-15&endDate=2022-04-22`

## Building

```
$ go install # or build I guess  
```

or

```
$ docker built -t mlb-rss .
```

for a container (scratch + binary + certificates)

## Usage

This program implements 2 subcommands: `generate` and `serve`

1. Generate the RSS content with `generate`:

  ```
  $ mlb-rss generate ./mlb-rss.xml
  ```

2. Serve that file over HTTP:

  ```
  $ mlb-rss serve ./mlb-rss.xml
  ```

3. Point your feed reader at port `8080`

4. Run Step 1 however often you want the feed to be updated,
   using a cron job or similar. It currently does not inspect
   the time of last generation, so every `generate` invocation
   will add a new entry. I only run it once a day anyways.
  
## My deployment

My personal server runs the containerized version with podman,
where the `serve` process is a normal systemd service, and the
`generate` process is a `OneShot` service activated by a timer,
writing to the host filesystem (rather than a volume, for easy
`cat` debugging). [Miniflux] is the feed reader it is known to
work with.

[Miniflux]: https://github.com/miniflux/miniflux
