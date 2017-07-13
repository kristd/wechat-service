package main

import "encoding/xml"

// XmlConfig: web api xml response struct
type XmlConfig struct {
    XMLName     xml.Name `xml:"error"`
    Ret         int      `xml:"ret"`
    Message     string   `xml:"message"`
    Skey        string   `xml:"skey"`
    Wxsid       string   `xml:"wxsid"`
    Wxuin       string   `xml:"wxuin"`
    PassTicket  string   `xml:"pass_ticket"`
    IsGrayscale int      `xml:"isgrayscale"`
}