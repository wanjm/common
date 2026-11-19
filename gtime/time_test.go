package gtime

import (
	"encoding/json"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// TestTimeJSON checks the value, pointer, and slice forms travel as Unix milliseconds and read back.
// A nil pointer is sent as null and the zero Time as 0.
func TestTimeJSON(t *testing.T) {
	type clock struct {
		Start Time   `json:"start"`
		End   *Time  `json:"end"`
		Marks []Time `json:"marks"`
	}
	end := FromMillis(1700000005000)
	data, err := json.Marshal(clock{Start: FromMillis(1700000000000), End: &end, Marks: []Time{FromMillis(1700000001000), {}}})
	if err != nil {
		t.Fatal(err)
	}
	if want := `{"start":1700000000000,"end":1700000005000,"marks":[1700000001000,0]}`; string(data) != want {
		t.Fatalf("json = %s, want %s", data, want)
	}

	var back clock
	if err := json.Unmarshal([]byte(`{"start":1700000000000,"end":null,"marks":[1700000001000,"0"]}`), &back); err != nil {
		t.Fatal(err)
	}
	if back.Start.Millis() != 1700000000000 || back.End != nil || back.Marks[0].Millis() != 1700000001000 || !back.Marks[1].IsZero() {
		t.Fatalf("decoded = %+v", back)
	}
	if err := json.Unmarshal([]byte(`{"start":"2023-11-14"}`), &back); err == nil {
		t.Fatal("a date string should be rejected")
	}
}

// TestTimeParam checks the gin query binding hook.
func TestTimeParam(t *testing.T) {
	var v Time
	if err := v.UnmarshalParam("1700000000000"); err != nil || v.Millis() != 1700000000000 {
		t.Fatalf("UnmarshalParam = %d, %v", v.Millis(), err)
	}
	if err := v.UnmarshalParam(""); err != nil || !v.IsZero() {
		t.Fatalf("empty param = %v, %v", v, err)
	}
}

// TestTimeSQL checks Value keeps a native time.Time and Scan accepts time and text forms.
func TestTimeSQL(t *testing.T) {
	at := Of(time.UnixMilli(1700000000000))
	v, err := at.Value()
	if got, ok := v.(time.Time); err != nil || !ok || !got.Equal(at.Time) {
		t.Fatalf("Value = %#v, %v", v, err)
	}
	var s Time
	for _, src := range []any{at.Time, []byte("2023-11-14 22:13:20"), "2023-11-14"} {
		if err := s.Scan(src); err != nil || s.IsZero() {
			t.Fatalf("Scan(%#v) = %v, %v", src, s, err)
		}
	}
	if err := s.Scan(nil); err != nil || !s.IsZero() {
		t.Fatalf("Scan(nil) = %v, %v", s, err)
	}
	if err := s.Scan("0000-00-00 00:00:00"); err != nil || !s.IsZero() {
		t.Fatalf("Scan(zero date) = %v, %v", s, err)
	}
}

// TestTimeBSON checks the instant is stored as a BSON date and reads back.
func TestTimeBSON(t *testing.T) {
	at := FromMillis(1700000000000)
	doc, err := bson.Marshal(bson.M{"at": at})
	if err != nil {
		t.Fatal(err)
	}
	if typ := bson.Raw(doc).Lookup("at").Type; typ != bson.TypeDateTime {
		t.Fatalf("bson type = %v, want datetime", typ)
	}
	var out struct {
		At Time `bson:"at"`
	}
	if err := bson.Unmarshal(doc, &out); err != nil || out.At.Millis() != at.Millis() {
		t.Fatalf("round trip = %d, %v", out.At.Millis(), err)
	}
}
