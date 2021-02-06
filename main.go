package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"cloud.google.com/go/logging/logadmin"
	"github.com/fatih/color"
	"github.com/monzo/slog"
	"google.golang.org/api/iterator"
)

var (
	project = flag.String("p", "", "your project ID")
	query = flag.String("q", "", "log query")
	lookback = flag.String("d", "1h", "lookback duration, eg '1h' or '30s'" )
	follow = flag.Bool("f", false, "follow the logs, like tail -f")
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
func run() error {
	flag.Parse()
	if *project == "" {
		return fmt.Errorf("set a project with -p")
	}
	ctx := context.Background()
	entries := make(chan slog.Event, 1024)
	go func() {
		seen := make(map[string]struct{})
		for e := range entries {
			if _, ok := seen[e.Id]; ok {
				continue
			}
			seen[e.Id] = struct{}{}
			Log(e)
		}
		os.Exit(0)
	}()
	client, err := logadmin.NewClient(ctx, *project)
	if err != nil {
		return err
	}
	err = poll(ctx,client,entries)
	if err != nil {
		return err
	}
	if !*follow {
		return nil
	}
	t := time.NewTicker(5 * time.Second)
	for range t.C {
		err = poll(ctx,client,entries)
		if err != nil {
			return err
		}
	}
	return nil
}

func poll(ctx context.Context, client *logadmin.Client, entries chan slog.Event) error {
	d, err := time.ParseDuration(*lookback)
	if err != nil {
		return err
	}
	window := time.Now().Add(d*-1)
	t := window.Format(time.RFC3339) // Logging API wants timestamps in RFC 3339 format.
	q := fmt.Sprintf(`timestamp > "%s" %s`, t, *query)
	it := client.Entries(ctx, logadmin.Filter(q))
	for {
		entry, err := it.Next()
		if err == iterator.Done {
			if !*follow {
				close(entries)
			}
			return nil
		} else if err != nil {
			return err
		}
		if entry.Payload != nil {
			entries <- slog.Event{
				Id: entry.InsertID,
				Timestamp: entry.Timestamp,
				Severity: slogSeverity(entry.Severity.String()),
				Message: fmt.Sprintf("%s", entry.Payload),
			}
		}
	}
}

func slogSeverity(s string) slog.Severity {
	switch s {
	case "Debug":
		return slog.DebugSeverity
	case "Info":
		return slog.InfoSeverity
	case "Warning":
		return slog.WarnSeverity
	case "Error":
		return slog.ErrorSeverity
	default:
		return 0
	}
}

func Log(evs ...slog.Event) {
	for _, e := range evs {
		switch e.Severity {
		case slog.TraceSeverity:
			fmt.Printf( "%s %s\n", color.WhiteString("%s TRC", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.DebugSeverity:
			fmt.Printf( "%s %s\n", color.CyanString("%s DBG", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.InfoSeverity:
			fmt.Printf( "%s %s\n", color.BlueString("%s INF", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.WarnSeverity:
			fmt.Printf("%s %s\n", color.YellowString("%s WRN", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.ErrorSeverity:
			fmt.Printf("%s %s\n", color.RedString("%s ERR", e.Timestamp.Format("15:04:05.000")), e.Message)
		case slog.CriticalSeverity:
			fmt.Printf("%s %s\n", color.RedString("%s CRT", e.Timestamp.Format("15:04:05.000")), e.Message)
		}
	}
}
