package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/rahmadafandi/clipboard-manager/internal/storage"
	"github.com/spf13/cobra"
	"golang.design/x/clipboard"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch clipboard for changes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Starting clipboard watcher...")
		
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

		for {
			select {
			case data := <-chText:
				text := string(data)
				if text == "" {
					continue
				}
				fmt.Println("Detected text copy")
				
				// Read current items to check duplicate
				items, _ := s.Load()
				if len(items) > 0 {
					last := items[len(items)-1]
					if last.Type == storage.Text && last.TextContent == text {
						continue
					}
				}
				
				items = append(items, storage.ClipItem{
					Type:        storage.Text,
					TextContent: text,
					Timestamp:   time.Now(),
				})
				
				// Limit
				if len(items) > 50 {
					items = items[len(items)-50:]
				}
				
				s.Save(items)

			case data := <-chImage:
				if len(data) == 0 {
					continue
				}
				fmt.Println("Detected image copy")
				
				items, _ := s.Load()
				// Simple duplicate check (length comparison for now to be fast)
				if len(items) > 0 {
					last := items[len(items)-1]
					if last.Type == storage.Image && len(last.ImageData) == len(data) {
						continue
					}
				}

				items = append(items, storage.ClipItem{
					Type:      storage.Image,
					ImageData: data,
					Timestamp: time.Now(),
				})
				
				if len(items) > 50 {
					items = items[len(items)-50:]
				}
				
				s.Save(items)
			}
		}
	},
}
