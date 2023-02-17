package main

import (
	"log"

	"github.com/go-redis/redis"
	"github.com/spf13/cobra"
)

type Options struct {
	SourceAddress       []string
	SourcePassword      string
	DestinationPassword string
	DestinationAddress  string
	DestinationDB       int
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

		sourceKeys := srcClient.Keys("*")
		keys, err := sourceKeys.Result()
		if err != nil {
			return err
		}

		log.Printf("Keys got %d keys\n", len(keys))
		log.Printf("Keys %+v\n", keys)

		successCount := 0

		for _, k := range keys {
			getResult := srcClient.Get(k)
			v, err := getResult.Result()
			if err != nil {
				return err
			}

			ttlResult := srcClient.TTL(k)
			ttl, err := ttlResult.Result()
			if err != nil {
				return err
			}

			ret := dstClient.Set(k, v, ttl)
			_, err = ret.Result()
			if err != nil {
				return err
			}

			successCount++
		}

		log.Printf("Successed count %d\n", successCount)
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

	rootCmd.PersistentFlags().StringVarP(&options.DestinationAddress, "dest-address", "", "", "destination addresses")
	rootCmd.PersistentFlags().StringVarP(&options.DestinationPassword, "dest-password", "", "", "destination password")
	rootCmd.PersistentFlags().IntVarP(&options.DestinationDB, "dest-db", "", 0, "destination db")
}
