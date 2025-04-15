package main

import (
	"fmt"
	"log"
	"time"

	report "github.com/oliverpool/go-dmarc-report"
	"github.com/spf13/pflag"
)

var (
	programStartTime = time.Now()
	programStartDate = time.Date(
		programStartTime.Year(),
		programStartTime.Month(),
		programStartTime.Day(),
		0, 0, 0, 0, time.Local,
	)
	dateFormat = "2006-01-02"
)

var (
	flagImapServer   = pflag.StringP("imap-server", "s", "", "IMAP4 server:port")
	flagImapUser     = pflag.StringP("imap-user", "u", "", "IMAP4 user name")
	flagImapPassword = pflag.StringP("imap-password", "p", "", "IMAP4 password")
	flagImapSince    = pflag.StringP("imap-since", "S", programStartDate.AddDate(0, 0, 0).Format(dateFormat), "Starting date for email report fetching")
	flagImapBefore   = pflag.StringP("imap-before", "B", programStartDate.AddDate(0, 0, 1).Format(dateFormat), "Ending date for email report fetching")
	flagFilename     = pflag.StringP("filename", "f", "", "File name. If specified, IMAP4 is not used")
)

func printAggregate(agg *report.Aggregate) {
	fmt.Println("Report from:")
	fmt.Printf("  org                : %s\n", agg.Metadata.OrgName)
	fmt.Printf("  email              : %s\n", agg.Metadata.Email)
	fmt.Printf("  extra contact info : %s\n", agg.Metadata.ExtraContactInfo)
	fmt.Printf("  report ID          : %s\n", agg.Metadata.ReportID)
	fmt.Printf("  from date          : %s\n", agg.Metadata.DateRange.Begin)
	fmt.Printf("  to date            : %s\n", agg.Metadata.DateRange.End)
	fmt.Println()
	fmt.Printf("Found %d record(s) in report\n", len(agg.Records))
	for idx, r := range agg.Records {
		fmt.Printf("% 3d) src=%s\n     from=%s\n     count=%d\n     success=%v\n     dkim_aligned=%+v\n     spf_aligned=%+v\n", idx+1, r.Row.SourceIP, r.Identifiers.HeaderFrom, r.Row.Count, r.FinalDispositionSuccess(), r.DKIMAligned(), r.SPFAligned())
	}
}

func main() {
	pflag.CommandLine.SortFlags = false
	pflag.Parse()

	if *flagFilename != "" {
		report := Report{
			Name: *flagFilename,
		}
		agg, err := parseReport(&report)
		if err != nil {
			log.Fatalf("Failed to parse report: %v", err)
		}
		printAggregate(agg)
	} else {
		since, err := time.Parse(dateFormat, *flagImapSince)
		if err != nil {
			log.Fatalf("Invalid date for --imap-since")
		}
		before, err := time.Parse(dateFormat, *flagImapBefore)
		if err != nil {
			log.Fatalf("Invalid date for --imap-before")
		}
		reports, err := getReportsViaIMAP4(*flagImapServer, *flagImapUser, *flagImapPassword, since, before)
		if err != nil {
			log.Fatalf("Failed to get reports via IMAP4: %v", err)
		}
		if len(reports) == 0 {
			log.Printf("No reports found in the requested time range")
			return
		}
		for _, report := range reports {
			agg, err := parseReport(report)
			if err != nil {
				log.Fatalf("Failed to parse report %s: %v", report.Name, err)
			}
			printAggregate(agg)
		}
	}
}
