package fusis

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Sirupsen/logrus"
)

//CustomLogFormatter custom formatter
type CustomLogFormatter struct{}

//Format Formats a logger entry
func (f *CustomLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}

	b := &bytes.Buffer{}

	fmt.Fprintf(b, "%s [%s] %-44s ", entry.Time.Format("2006/01/02 15:04:05"), strings.ToUpper(entry.Level.String()), entry.Message)

	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, "%s=%+v", k, v)
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *CustomLogFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {

	b.WriteString(key)
	b.WriteByte('=')

	switch value := value.(type) {
	case string:
		if !needsQuoting(value) {
			b.WriteString(value)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	case error:
		errmsg := value.Error()
		if !needsQuoting(errmsg) {
			b.WriteString(errmsg)
		} else {
			fmt.Fprintf(b, "%q", value)
		}
	default:
		fmt.Fprint(b, value)
	}

	b.WriteByte(' ')
}

func needsQuoting(text string) bool {
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.') {
			return true
		}
	}
	return false
}
