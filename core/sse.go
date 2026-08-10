package core

import (
	"bufio"
	"io"
	"strings"
)

func ScanSSE(reader *bufio.Reader, handle func(event, data string) bool) error {
	var (
		event string
		data  strings.Builder
	)

	dispatch := func() bool {
		if data.Len() == 0 {
			event = ""
			return true
		}
		ok := handle(event, data.String())
		event = ""
		data.Reset()
		return ok
	}

	for {
		raw, readErr := reader.ReadString('\n')

		if raw != "" {
			line := strings.TrimRight(raw, "\r\n")
			switch {
			case line == "":
				if !dispatch() {
					return nil
				}
			case strings.HasPrefix(line, ":"):
				// * comment / keep-alive
			default:
				field, value, _ := strings.Cut(line, ":")
				value = strings.TrimPrefix(value, " ")
				switch field {
				case "event":
					event = value
				case "data":
					if data.Len() > 0 {
						data.WriteByte('\n')
					}
					data.WriteString(value)
				}
			}
		}

		if readErr != nil {
			if readErr == io.EOF {
				dispatch()
				return nil
			}
			return readErr
		}
	}
}
