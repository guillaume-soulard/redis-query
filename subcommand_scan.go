package main

import (
	"context"
	"errors"
	"fmt"
)

type ScanSubCommand struct{}

func (s ScanSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Scan.Cmd.Happened()
}

func (s ScanSubCommand) Execute(parameters *Parameters) (err error) {
	client := connectToRedis(parameters.Scan.Connect)
	cursor := uint64(0)
	var result []string
	var keyType string
	key := *parameters.Scan.KeyToScan
	if key != "" {
		if keyType, err = client.Type(context.Background(), key).Result(); err != nil {
			return err
		}
	}
	rowNumber := 0
	for {
		if keyType == "set" {
			if result, cursor, err = client.SScan(context.Background(), key, cursor, *parameters.Scan.Pattern, int64(*parameters.Scan.Count)).Result(); err != nil {
				return err
			}
		} else if keyType == "zset" {
			if result, cursor, err = client.ZScan(context.Background(), key, cursor, *parameters.Scan.Pattern, int64(*parameters.Scan.Count)).Result(); err != nil {
				return err
			}
		} else if keyType == "hash" {
			if result, cursor, err = client.HScan(context.Background(), key, cursor, *parameters.Scan.Pattern, int64(*parameters.Scan.Count)).Result(); err != nil {
				return err
			}
		} else if keyType == "" {
			if *parameters.Scan.Type != "" {
				if result, cursor, err = client.ScanType(context.Background(), cursor, *parameters.Scan.Pattern, int64(*parameters.Scan.Count), *parameters.Scan.Type).Result(); err != nil {
					return err
				}
			} else {
				if result, cursor, err = client.Scan(context.Background(), cursor, *parameters.Scan.Pattern, int64(*parameters.Scan.Count)).Result(); err != nil {
					return err
				}
			}
		} else {
			err = errors.New(fmt.Sprintf("Unable to scan key type: %s", keyType))
			return err
		}
		for _, k := range result {
			if *parameters.Scan.Limit > 0 && rowNumber >= *parameters.Scan.Limit {
				break
			}
			formatIfNeededAndPrint(&rowNumber, "", k, &parameters.Scan.Format)
		}
		if cursor == 0 || (*parameters.Scan.Limit > 0 && rowNumber >= *parameters.Scan.Limit) {
			break
		}
	}
	return err
}
