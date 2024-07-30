package main

import (
	"context"
	"errors"
	"fmt"
)

func scan(params Parameters) {
	client := connectToRedis(params.Scan.Connect)
	cursor := uint64(0)
	var err error
	var result []string
	var keyType string
	key := *params.Scan.KeyToScan
	if key != "" {
		if keyType, err = client.Type(context.Background(), key).Result(); err != nil {
			PrintErrorAndExit(err)
		}
	}
	rowNumber := 0
	for {
		if keyType == "set" {
			if result, cursor, err = client.SScan(context.Background(), key, cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else if keyType == "zset" {
			if result, cursor, err = client.ZScan(context.Background(), key, cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else if keyType == "hash" {
			if result, cursor, err = client.HScan(context.Background(), key, cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
				PrintErrorAndExit(err)
			}
		} else if keyType == "" {
			if *params.Scan.Type != "" {
				if result, cursor, err = client.ScanType(context.Background(), cursor, *params.Scan.Pattern, int64(*params.Scan.Count), *params.Scan.Type).Result(); err != nil {
					PrintErrorAndExit(err)
				}
			} else {
				if result, cursor, err = client.Scan(context.Background(), cursor, *params.Scan.Pattern, int64(*params.Scan.Count)).Result(); err != nil {
					PrintErrorAndExit(err)
				}
			}
		} else {
			PrintErrorAndExit(errors.New(fmt.Sprintf("Unable to scan key type: %s", keyType)))
		}
		for _, key = range result {
			if *params.Scan.Limit > 0 && rowNumber >= *params.Scan.Limit {
				break
			}
			formatIfNeededAndPrint(&rowNumber, "", key, &params.Scan.Format)
		}
		if cursor == 0 || (*params.Scan.Limit > 0 && rowNumber >= *params.Scan.Limit) {
			break
		}
	}
}
