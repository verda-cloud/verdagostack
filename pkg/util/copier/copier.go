// Copyright 2026 Verda Cloud Oy
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package copier provides deep-copy helpers with built-in time.Time ↔
// timestamppb.Timestamp converters. See README.md for examples.
package copier

import (
	"errors"
	"time"

	"github.com/jinzhu/copier"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// TypeConverters returns copier type converters for time.Time <-> timestamppb.Timestamp.
func TypeConverters() []copier.TypeConverter {
	return []copier.TypeConverter{
		{
			SrcType: time.Time{},
			DstType: &timestamppb.Timestamp{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(time.Time)
				if !ok {
					return nil, errors.New("source type not matching")
				}
				return timestamppb.New(s), nil
			},
		},
		{
			SrcType: &timestamppb.Timestamp{},
			DstType: time.Time{},
			Fn: func(src interface{}) (interface{}, error) {
				s, ok := src.(*timestamppb.Timestamp)
				if !ok {
					return nil, errors.New("source type not matching")
				}
				return s.AsTime(), nil
			},
		},
	}
}

// CopyWithConverters performs a deep copy from src to dst using the built-in
// time converters plus any additional converters provided.
func CopyWithConverters(to any, from any, converters ...copier.TypeConverter) error {
	converters = append(TypeConverters(), converters...)
	return copier.CopyWithOption(to, from, copier.Option{IgnoreEmpty: true, DeepCopy: true, Converters: converters})
}

// Copy performs a shallow copy from src to dst.
func Copy(to any, from any) error {
	return copier.Copy(to, from)
}
