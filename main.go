package main

import (
	"errors"
	"log"
	"strings"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

type Options struct {
	SourceAddress       []string
	SourcePassword      string
	DestinationPassword string
	DestinationAddress  string
	DestinationDB       int
	Keys                string
}

var options Options
var rootCmd = cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("%+v\n\n", options)
		dstClient := redis.NewClient(&redis.Options{
			Addr:     options.DestinationAddress,
			Password: options.DestinationPassword,
			DB:       options.DestinationDB,
		})

		srcClient := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    options.SourceAddress,
			Password: options.SourcePassword,
		})

		if err := srcClient.ForEachMaster(func(client *redis.Client) error {
			sourceKeys := client.Keys(options.Keys)
			keys, err := sourceKeys.Result()
			if err != nil {
				panic(err)
			}

			nameResult := client.ClientGetName()

			log.Printf("[%s] Keys got %d keys\n", nameResult.Val(), len(keys))
			log.Printf("[%s]\nKeys %+v\n", nameResult.Val(), keys)

			successCount := 0

			for _, k := range keys {
				getResult := client.Get(k)

				v, err := getResult.Result()

				if err != nil {
					if errors.Is(err, redis.Nil) {
						log.Printf("key `%s` is nil", k)
						continue
					} else {
						if strings.Contains(err.Error(), "MOVED") {
							continue
						} else {
							panic(err)
						}
					}
				}

				ttlResult := client.TTL(k)
				ttl, err := ttlResult.Result()
				if err != nil {
					panic(err)
				}

				ret := dstClient.Set(k, v, ttl)
				_, err = ret.Result()
				if err != nil {
					panic(err)
				}

				successCount++
			}

			log.Printf("Successed count %d\n", successCount)

			return nil
		}); err != nil {
			panic(err)
		}

		return nil
	},
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringArrayVarP(&options.SourceAddress, "source-address", "", []string{}, "source addresses")
	rootCmd.PersistentFlags().StringVarP(&options.SourcePassword, "source-password", "", "", "source password")
	rootCmd.PersistentFlags().StringVarP(&options.Keys, "source-key", "", "*", "source keys")

	rootCmd.PersistentFlags().StringVarP(&options.DestinationAddress, "dest-address", "", "", "destination addresses")
	rootCmd.PersistentFlags().StringVarP(&options.DestinationPassword, "dest-password", "", "", "destination password")
	rootCmd.PersistentFlags().IntVarP(&options.DestinationDB, "dest-db", "", 0, "destination db")
}
