// Package gtime holds Time, an instant that travels over HTTP JSON as Unix milliseconds.
package gtime

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

// Time is an instant that travels over HTTP JSON as Unix milliseconds.
// It embeds time.Time, so every time.Time method stays available; the zero Time is sent as 0.
// gos reads the wire annotation and generates Date / DateTime for it on the clients.
// @gos wire=millis
type Time struct {
	time.Time
}

// Now returns the current instant as a Time.
func Now() Time {
	return Time{time.Now()}
}

// Of wraps a time.Time.
func Of(t time.Time) Time {
	return Time{t}
}

// FromMillis builds a Time from Unix milliseconds.
// 0 gives the zero Time, not the Unix epoch.
func FromMillis(ms int64) Time {
	if ms == 0 {
		return Time{}
	}
	return Time{time.UnixMilli(ms)}
}

// Millis returns the instant as Unix milliseconds.
// The zero Time gives 0.
func (t Time) Millis() int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}

// MarshalJSON writes the instant as a JSON number of Unix milliseconds.
func (t Time) MarshalJSON() ([]byte, error) {
	return strconv.AppendInt(nil, t.Millis(), 10), nil
}

// UnmarshalJSON reads Unix milliseconds given as a number or a numeric string.
// null, "" and 0 give the zero Time.
func (t *Time) UnmarshalJSON(data []byte) error {
	return t.parseMillis(strings.Trim(string(data), "\""))
}

// UnmarshalParam lets gin bind query and form values given as Unix milliseconds.
func (t *Time) UnmarshalParam(param string) error {
	return t.parseMillis(param)
}

// parseMillis sets t from a decimal Unix millisecond string.
func (t *Time) parseMillis(s string) error {
	if s == "" || s == "null" {
		*t = Time{}
		return nil
	}
	ms, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("gtime: %q is not Unix milliseconds", s)
	}
	*t = FromMillis(ms)
	return nil
}

// GormDataType tells gorm to treat the column as a time column.
func (Time) GormDataType() string {
	return "time"
}

// Value stores the instant as a SQL datetime, the same way time.Time is stored.
func (t Time) Value() (driver.Value, error) {
	return t.Time, nil
}

// Scan reads a SQL datetime, or its text form when the DSN has no parseTime=true.
func (t *Time) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*t = Time{}
	case time.Time:
		*t = Time{v}
	case []byte:
		return t.scanText(string(v))
	case string:
		return t.scanText(v)
	default:
		return fmt.Errorf("gtime: cannot scan %T into Time", src)
	}
	return nil
}

// scanText parses a MySQL DATETIME / DATE text value in local time.
func (t *Time) scanText(s string) error {
	if s == "" || strings.HasPrefix(s, "0000-00-00") {
		*t = Time{}
		return nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05.999999999", "2006-01-02", time.RFC3339Nano} {
		if v, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			*t = Time{v}
			return nil
		}
	}
	return fmt.Errorf("gtime: cannot parse %q as a datetime", s)
}

// MarshalBSONValue stores the instant as a BSON date, the same way time.Time is stored.
func (t Time) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(t.Time)
}

// UnmarshalBSONValue reads a BSON date; null gives the zero Time.
func (t *Time) UnmarshalBSONValue(typ bsontype.Type, data []byte) error {
	if typ == bson.TypeNull {
		*t = Time{}
		return nil
	}
	var v time.Time
	if err := (bson.RawValue{Type: typ, Value: data}).Unmarshal(&v); err != nil {
		return err
	}
	*t = Time{v}
	return nil
}
