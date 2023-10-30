package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

func getReportsViaIMAP4(server, user, password string, since, before time.Time) ([]*Report, error) {
	client, err := imapclient.DialTLS(server, nil)
	if err != nil {
		return nil, fmt.Errorf("connect failed: %w", err)
	}
	defer client.Close()

	cmd := client.Login(user, password)
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// select the mailbox
	if _, err := client.Select("INBOX", nil).Wait(); err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}

	searchData, err := client.Search(&imap.SearchCriteria{
		Since:  since,
		Before: before,
		Header: []imap.SearchCriteriaHeaderField{
			{
				Key:   "Subject",
				Value: "Report Domain: ",
			},
		},
	}, nil).Wait()
	if err != nil {
		return nil, fmt.Errorf("IMAP4 search failed: %w", err)
	}
	if len(searchData.All) == 0 {
		return nil, nil
	}
	fetchOptions := &imap.FetchOptions{
		Flags:    true,
		Envelope: true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierText},
		},
	}
	msgs, err := client.Fetch(searchData.All, fetchOptions).Collect()
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	reports := make([]*Report, 0)
	for _, msg := range msgs {
		for section, body := range msg.BodySection {
			if section.Specifier == imap.PartSpecifierText {
				data, err := base64.StdEncoding.DecodeString(string(body))
				if err != nil {
					return nil, fmt.Errorf("base64 decode failed: %w", err)
				}
				// FIXME: I am not sure how to get the MIME type from the email
				// headers, so.. MIME type detection :(
				mimeType := http.DetectContentType(data)
				reports = append(reports, &Report{
					Data: data,
					Type: mimeType,
				})
			}
		}
	}

	return reports, nil
}
