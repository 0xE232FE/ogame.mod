package utils

import (
	"bytes"
	"compress/gzip"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseInt ...
func ParseInt(val string) int64 {
	val = strings.Replace(val, ".", "", -1)
	val = strings.Replace(val, ",", "", -1)
	val = strings.TrimSpace(val)
	return DoParseI64(val)
}

func ToInt(buf []byte) (n int) {
	for _, v := range buf {
		n = n*10 + int(v-'0')
	}
	return
}

// I64Ptr returns a pointer to int64
func I64Ptr(v int64) *int64 {
	return &v
}

// MinInt returns the minimum int64 value
func MinInt(vals ...int64) int64 {
	min := vals[0]
	for _, num := range vals {
		if num < min {
			min = num
		}
	}
	return min
}

// MaxInt returns the minimum int64 value
func MaxInt(vals ...int64) int64 {
	max := vals[0]
	for _, num := range vals {
		if num > max {
			max = num
		}
	}
	return max
}

// Clamp ensure the value is within a range
func Clamp(val, min, max int64) int64 {
	val = MinInt(val, max)
	val = MaxInt(val, min)
	return val
}

func ParseI64(v string) (out int64, err error) {
	return strconv.ParseInt(v, 10, 64)
}

func DoParseI64(v string) (out int64) {
	out, _ = ParseI64(v)
	return
}

type Ints interface {
	~int64 | ~int
}

// FI64 formats any int types to string
func FI64[T Ints](v T) string {
	return strconv.FormatInt(int64(v), 10)
}

func DoCastF64(v any) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}

func DoCastStr(v any) string {
	if str, ok := v.(string); ok {
		return str
	}
	return ""
}

func GetNbr(doc *goquery.Document, name string) int64 {
	div := doc.Find("div." + name)
	level := div.Find("span.level")
	level.Children().Remove()
	return ParseInt(level.Text())
}

func GetNbrShips(doc *goquery.Document, name string) int64 {
	div := doc.Find("div." + name)
	title := div.AttrOr("title", "")
	if title == "" {
		title = div.Find("a").AttrOr("title", "")
	}
	m := regexp.MustCompile(`.+\(([\d.,]+)\)`).FindStringSubmatch(title)
	if len(m) != 2 {
		return 0
	}
	return ParseInt(m[1])
}

func ReadBody(resp *http.Response) (respContent []byte, err error) {
	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		var err error
		reader, err = gzip.NewReader(buf)
		if err != nil {
			return []byte{}, err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}
	by, err := ioutil.ReadAll(reader)
	if err != nil {
		return []byte{}, err
	}
	return by, nil
}

type Equalable[T any] interface {
	Equal(other T) bool
}

func InArray[T Equalable[T]](needle T, haystack []T) bool {
	for _, el := range haystack {
		if needle.Equal(el) {
			return true
		}
	}
	return false
}

// Random generates a number between min and max inclusively
func Random(min, max int64) int64 {
	if min == max {
		return min
	}
	if max < min {
		min, max = max, min
	}
	return rand.Int63n(max-min+1) + min
}

// RandDuration generates random duration
func RandDuration(min, max time.Duration) time.Duration {
	n := Random(min.Nanoseconds(), max.Nanoseconds())
	return time.Duration(n) * time.Nanosecond
}

func randDur(min, max int64, dur time.Duration) time.Duration {
	return RandDuration(time.Duration(min)*dur, time.Duration(max)*dur)
}

// RandMs generates random duration in milliseconds
func RandMs(min, max int64) time.Duration {
	return randDur(min, max, time.Millisecond)
}
