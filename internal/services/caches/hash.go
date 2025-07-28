package caches

import (
	"context"
	"fmt"
	initializer "go-cache-aside-service/init"
	"log"
	"reflect"
	"time"
)

type HashCollection struct {
	KeyName    string `json:"keyname"`
	Collection []Hash `json:"collection"`
}

type Hash struct {
	Key         string   `json:"key"`
	Fields      []string `json:"fields"`
	FieldsTTL   []int    `json:"fields_ttl"`
	Values      []any    `json:"values"`
	FieldValues []any    `json:"fields_values"`
}

// ::: -> Hash Collection
func (hc *HashCollection) Add(ttl time.Duration) error {
	rdb, err := initializer.GetRedisDB()
	if err != nil {
		return err
	}

	ctx := context.Background()
	pipe := rdb.Pipeline()
	for _, hash := range hc.Collection {
		pipe.HSet(ctx, hash.Key, hash.FieldValues...)
		if ttl.Nanoseconds() != 0 {
			pipe.HExpire(ctx, hash.Key, ttl, hash.Fields...)
		}
	}

	if _, errExec := pipe.Exec(ctx); errExec != nil {
		return errExec
	}
	return nil
}

// [:data] must be an Interface or Any
func ExtractToHash(keyPropName string, data any) HashCollection {
	var redis_hash_collection HashCollection

	dataType := reflect.ValueOf(data)

	// SLICE
	if dataType.Kind() == reflect.Slice {
		log.Println("Extract [slice] to Hash ...")
		for i := 0; i < dataType.Len(); i++ {
			element := dataType.Index(i).Interface()

			redis_hash := &Hash{}

			tElement := reflect.TypeOf(element)
			vElement := reflect.ValueOf(element)

			// SLICE -> MAP
			if tElement.Kind() == reflect.Map {
				mapKey := reflect.ValueOf(keyPropName)
				val := vElement.MapIndex(mapKey)

				if !val.IsValid() {
					panic("key prop name: invalid")
				} else {
					redis_hash.Key = fmt.Sprintf("%v", val.Interface())
				}

				ExtractMapToHash(element, redis_hash)
			}

			// SLICE -> STRUCT
			if tElement.Kind() == reflect.Struct {
				ExtractStructToHash(keyPropName, element, redis_hash)
			}

			redis_hash_collection.Collection = append(redis_hash_collection.Collection, *redis_hash)
		}
	}

	// STRUCT
	if dataType.Kind() == reflect.Struct {
		log.Println("Extract [struct] to Hash ...")
		structValue := dataType.Interface()

		redis_hash := &Hash{}

		ExtractStructToHash(keyPropName, structValue, redis_hash)

		redis_hash_collection.Collection = append(redis_hash_collection.Collection, *redis_hash)
	}

	// MAP
	if dataType.Kind() == reflect.Map {
		log.Println("Extract [map] to Hash ...")
		mapValue := dataType.Interface()

		redis_hash := &Hash{}

		mapKey := reflect.ValueOf(keyPropName)
		val := dataType.MapIndex(mapKey)

		if !val.IsValid() {
			panic("key prop name: invalid")
		} else {
			redis_hash.Key = fmt.Sprintf("%v", val.Interface())
		}

		ExtractMapToHash(mapValue, redis_hash)

		redis_hash_collection.Collection = append(redis_hash_collection.Collection, *redis_hash)
	}

	redis_hash_collection.KeyName = keyPropName
	return redis_hash_collection
}

// :::Map Responsibility
func ExtractMapToHash(data any, redis_hash *Hash) {
	iter := reflect.ValueOf(data).MapRange()

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		if val.Elem().Kind() == reflect.Map {
			MapIterations(key, val.Interface(), redis_hash)
			continue
		}

		redis_hash.Fields = append(redis_hash.Fields, fmt.Sprintf("%v", key.Interface()))
		redis_hash.Values = append(redis_hash.Values, val.Interface())
		redis_hash.FieldValues = append(redis_hash.FieldValues, key.Interface(), val.Interface())
	}
}
func MapIterations(keyPrefix any, val any, redis_hash *Hash) {
	iter := reflect.ValueOf(val).MapRange()

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		composeKey := fmt.Sprintf("%v.%v", keyPrefix, key)

		if val.Elem().Kind() == reflect.Map {
			MapIterations(composeKey, val.Interface(), redis_hash)
			continue
		}

		redis_hash.Fields = append(redis_hash.Fields, composeKey)
		redis_hash.Values = append(redis_hash.Values, val.Interface())
		redis_hash.FieldValues = append(redis_hash.FieldValues, composeKey, val.Interface())
	}
}

// :::Struct Responsibility
func ExtractStructToHash(keyPropName string, data any, redis_hash *Hash) {
	v := reflect.ValueOf(data)

	for i := 0; i < v.Type().NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("cache")
		if tag == "" {
			tag = v.Type().Field(i).Name
		}

		if tag == keyPropName {
			redis_hash.Key = fmt.Sprintf("%v", v.Field(i).Interface())
		}

		if tm, ok := v.Field(i).Interface().(time.Time); ok { // verify time.Time and just get that formatted value, otherwise will got into decision struct and panic
			redis_hash.Fields = append(redis_hash.Fields, tag)
			redis_hash.Values = append(redis_hash.Values, tm)
			redis_hash.FieldValues = append(redis_hash.FieldValues, tag, tm)
			continue
		}

		if v.Field(i).Kind() == reflect.Struct {
			StructIterations(tag, v.Field(i).Interface(), redis_hash)
			continue
		}

		redis_hash.Fields = append(redis_hash.Fields, tag)
		redis_hash.Values = append(redis_hash.Values, v.Field(i).Interface())
		redis_hash.FieldValues = append(redis_hash.FieldValues, tag, v.Field(i).Interface())
	}
}
func StructIterations(keyPrefix any, val any, redis_hash *Hash) {
	v := reflect.ValueOf(val)

	for i := 0; i < v.Type().NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("cache")
		if tag == "" {
			tag = v.Type().Field(i).Name
		}

		composeKey := fmt.Sprintf("%s.%s", keyPrefix, tag)

		if tm, ok := v.Field(i).Interface().(time.Time); ok { // verify time.Time and just get that formatted value, otherwise will got into decision struct and panic
			redis_hash.Fields = append(redis_hash.Fields, composeKey)
			redis_hash.Values = append(redis_hash.Values, tm)
			redis_hash.FieldValues = append(redis_hash.FieldValues, composeKey, tm)
			continue
		}

		if v.Field(i).Kind() == reflect.Struct {
			StructIterations(composeKey, v.Field(i).Interface(), redis_hash)
			continue
		}

		redis_hash.Fields = append(redis_hash.Fields, composeKey)
		redis_hash.Values = append(redis_hash.Values, v.Field(i).Interface())
		redis_hash.FieldValues = append(redis_hash.FieldValues, composeKey, v.Field(i).Interface())
	}
}
