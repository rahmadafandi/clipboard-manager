package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rahmadafandi/clipboard-manager/internal/config"
	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch clipboard for changes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting clipboard watcher...")

		cfg, _ := config.Load()

		err := clipboard.Init()
		if err != nil {
			panic(err)
		}

		s, err := storage.NewFileStorage()
		if err != nil {
			panic(err)
		}

		chText := clipboard.Watch(context.Background(), clipboard.FmtText)
		chImage := clipboard.Watch(context.Background(), clipboard.FmtImage)

		fmt.Printf("Saving history to: %s\n", s.FilePath)
		fmt.Printf("Max history: %d\n", cfg.MaxHistory)
		if cfg.AutoExpireHours > 0 {
			fmt.Printf("Auto-expire: %d hours\n", cfg.AutoExpireHours)
		}

		// Purge expired items on startup
		s.PurgeExpired()

		// Periodic expiry check
		if cfg.AutoExpireHours > 0 {
			go func() {
				ticker := time.NewTicker(10 * time.Minute)
				defer ticker.Stop()
				for range ticker.C {
					s.PurgeExpired()
				}
			}()
		}

		for {
			select {
			case data := <-chText:
				text := string(data)
				if text == "" {
					continue
				}
				fmt.Println("Detected text copy")

				item := storage.ClipItem{
					Type:        storage.Text,
					TextContent: text,
					Timestamp:   time.Now(),
				}

				if cfg.AutoExpireHours > 0 {
					exp := time.Now().Add(time.Duration(cfg.AutoExpireHours) * time.Hour)
					item.ExpiresAt = &exp
				}

				s.AppendWithLimit(item, cfg.MaxHistory)

			case data := <-chImage:
				if len(data) == 0 {
					continue
				}
				fmt.Println("Detected image copy")

				items, _ := s.Load()
				if len(items) > 0 {
					last := items[len(items)-1]
					if last.Type == storage.Image && len(last.ImageData) == len(data) {
						continue
					}
				}

				item := storage.ClipItem{
					Type:      storage.Image,
					ImageData: data,
					Timestamp: time.Now(),
				}

				if cfg.AutoExpireHours > 0 {
					exp := time.Now().Add(time.Duration(cfg.AutoExpireHours) * time.Hour)
					item.ExpiresAt = &exp
				}

				items = append(items, item)
				if len(items) > cfg.MaxHistory {
					items = items[len(items)-cfg.MaxHistory:]
				}
				s.Save(items)
			}
		}
	},
}
