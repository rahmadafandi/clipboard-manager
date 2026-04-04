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

		cfg, err := config.Load()
		if err != nil {
			fmt.Println("Warning: config load error, using defaults:", err)
		}

		if err := clipboard.Init(); err != nil {
			fmt.Println("Failed to init clipboard:", err)
			return
		}

		s, err := storage.NewFileStorage()
		if err != nil {
			fmt.Println("Failed to init storage:", err)
			return
		}

		chText := clipboard.Watch(context.Background(), clipboard.FmtText)
		chImage := clipboard.Watch(context.Background(), clipboard.FmtImage)

		fmt.Printf("Saving history to: %s\n", s.Path())
		fmt.Printf("Max history: %d\n", cfg.MaxHistory)
		if cfg.AutoExpireHours > 0 {
			fmt.Printf("Auto-expire: %d hours\n", cfg.AutoExpireHours)
		}

		s.PurgeExpired()

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
				if err := s.AppendWithLimit(item, cfg.MaxHistory); err != nil {
					fmt.Println("Error saving:", err)
				}

			case data := <-chImage:
				if len(data) == 0 {
					continue
				}
				fmt.Println("Detected image copy")
				item := storage.ClipItem{
					Type:      storage.Image,
					ImageData: data,
					Timestamp: time.Now(),
				}
				if cfg.AutoExpireHours > 0 {
					exp := time.Now().Add(time.Duration(cfg.AutoExpireHours) * time.Hour)
					item.ExpiresAt = &exp
				}
				if err := s.AppendWithLimit(item, cfg.MaxHistory); err != nil {
					fmt.Println("Error saving:", err)
				}
			}
		}
	},
}
