package main

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	log "github.com/cantara/bragi"
	cloud "github.com/cantara/nerthus/aws"
	"github.com/cantara/nerthus/aws/metadata"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}
	sess, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.AddError(err).Fatal("While creating aws session")
	}
	sess.Region = os.Getenv("region")

	var c cloud.AWS
	// Create an EC2 service client.
	c.NewEC2(sess)

	ids, err := metadata.GetAllServerIds(c.GetEC2())
	if err != nil {
		log.AddError(err).Fatal("while getting all server ids")
		return
	}
	log.Println(ids)
	ids, err = metadata.GetAllServersWithMetadataV1IDs(c.GetEC2())
	if err != nil {
		log.AddError(err).Fatal("while getting all server ids")
		return
	}
	log.Println(ids)
	for _, id := range ids {
		err = metadata.SetMetadataV2(id, c.GetEC2())
		if err != nil {
			log.AddError(err).Fatal("while setting server ", id, " metadata to version 2, ABORTING")
			return
		}
	}
}
