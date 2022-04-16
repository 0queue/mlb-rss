package rss

import "encoding/xml"

type Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Items       []Item   `xml:"item"`
}

type Item struct {
	Title       string       `xml:"title"`
	Link        string       `xml:"link"`
	Description *Description `xml:"description"`
	Guid        string       `xml:"guid"`
	// too lazy to make an xml deserializer for time that's rfc822
	PubDate string `xml:"pubDate"`
}

type Description struct {
	XMLName xml.Name `xml:"description"`
	Text    string   `xml:",cdata"`
}
