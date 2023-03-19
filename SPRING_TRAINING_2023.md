# Spring Training 2023

Redoing basically the whole thing for the 2023 season, now that I know
more about go, helix, etc.

Goals:
- no persistence, rather a small in memory cache refreshed daily
  with stable IDs so miniflux doesn't get understandably confused
- distroless/static instead of copying ssl certs and what not
- image to Github
- tests and a command to easily scrape real test data

