package helpers

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Path will return absolute path to folder
// that contains executable binary file.
func Path(file ...string) string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Error(err)
		return ""
	}

	if file != nil {
		return fmt.Sprintf("%s/%s", dir, filepath.Join(file...))
	}

	return dir
}

// timeNow is simple function to return pointer to time struct
func TimeNow() *time.Time {
	t := time.Now()
	return &t
}

// strincConversion converts string value to appropriate interface.
func stringConversion(in string) interface{} {
	if f64, err := strconv.ParseFloat(in, 64); err == nil {
		return f64
	}

	if i64, err := strconv.Atoi(in); err == nil {
		return i64
	}

	if in == "true" {
		return true
	}

	if in == "false" {
		return false
	}

	return in
}

// Convert reponse data to list, after this
// data will be Marshaled.
func CreateMapFromPipeChar(s string) map[string]interface{} {

	items := make(map[string]interface{})
	var listItems []interface{}
	pipeSplit := strings.Split(s, "|")

	for _, v := range pipeSplit {

		v = strings.TrimLeft(v, ",")

		split := strings.Split(v, ",")

		first := strings.Split(split[0], "=")

		// check if it is a list, we find out this by checking first value
		// if it is digit
		isList := strings.ToLower(first[0]) == "pool"

		if isList {
			fields := make(map[string]interface{})
			for _, j := range split {
				s := strings.Split(j, "=")
				fields[s[0]] = stringConversion(s[1])
			}
			listItems = append(listItems, fields)
		} else {
			for _, j := range split {
				s := strings.Split(j, "=")
				if len(s) < 2 {
					continue
				}
				items[s[0]] = stringConversion(s[1])
			}
		}
	}

	items["list"] = listItems

	return items
}

func BuildURL(parts []map[string]string) string {
	var buf strings.Builder
	for _, v := range parts {
		for a, b := range v {
			if buf.Len() > 0 {
				buf.WriteByte('&')
			}
			buf.WriteString(url.QueryEscape(a))
			buf.WriteByte('=')
			buf.WriteString(url.QueryEscape(b))
		}
	}

	return buf.String()
}
