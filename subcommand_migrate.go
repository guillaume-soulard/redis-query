package main

import (
	"context"
	"time"
)

type MigrateSubCommand struct{}

func (s MigrateSubCommand) Accept(parameters *Parameters) bool {
	return parameters.Migrate.Cmd.Happened()
}

func (s MigrateSubCommand) Execute(parameters *Parameters) (err error) {
	sourceClient := connectToRedis(parameters.Migrate.SourceConnect)
	targetClient := connectToRedis(parameters.Migrate.TargetConnect)
	count := int64(10)
	if parameters.Migrate.Count != nil && *parameters.Migrate.Count > 0 {
		count = int64(*parameters.Migrate.Count)
	}
	var cursor uint64
	var keys []string
	nbKeys := 0
	var dump string
	ttl := time.Duration(-1)
	if parameters.Migrate.Ttl != nil && *parameters.Migrate.Ttl >= -1 {
		ttl = time.Duration(*parameters.Migrate.Ttl)
	}
	for {
		if keys, cursor, err = sourceClient.Scan(context.Background(), cursor, *parameters.Migrate.SourcePattern, count).Result(); err != nil {
			return err
		}
		for _, key := range keys {
			if parameters.Migrate.Limit != nil && *parameters.Migrate.Limit > 0 && nbKeys >= *parameters.Migrate.Limit {
				break
			}
			if dump, err = sourceClient.Dump(context.Background(), key).Result(); err != nil {
				return err
			}
			ttl = -1
			if parameters.Migrate.Ttl == nil || *parameters.Migrate.Ttl == 0 {
				if ttl, err = sourceClient.TTL(context.Background(), key).Result(); err != nil {
					return err
				}
			}
			if parameters.Migrate.Replace != nil && *parameters.Migrate.Replace {
				if err = targetClient.RestoreReplace(context.Background(), key, ttl, dump).Err(); err != nil {
					if err.Error() == "BUSYKEY Target key name already exists." {
						err = nil
					} else {
						return err
					}
				}
			} else {
				if err = targetClient.Restore(context.Background(), key, ttl, dump).Err(); err != nil {
					if err.Error() == "BUSYKEY Target key name already exists." {
						err = nil
					} else {
						return err
					}
				}
			}
			nbKeys += 1
		}
		if parameters.Migrate.Limit != nil && *parameters.Migrate.Limit > 0 && nbKeys >= *parameters.Migrate.Limit {
			break
		}
		if cursor == 0 {
			break
		}
	}

	return err
}
