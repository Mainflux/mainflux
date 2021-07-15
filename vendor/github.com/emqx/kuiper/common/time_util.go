package common

import (
	"fmt"
	"github.com/benbjohnson/clock"
	"time"
)

//var (
//	formats = map[string]string{
//		"MMMM": "January",
//		"MMM":  "Jan",
//		"MM":   "01",
//		"M":    "1",
//		"YYYY": "2006",
//		"yyyy": "2006",
//		"YY":   "06",
//		"yy":   "06",
//		"G":    "AD",
//		"EEEE": "Monday",
//		"EEE":  "Mon",
//		"dd":   "02",
//		"d":    "2",
//		"HH":   "15",
//		"hh":   "03",
//		"h":    "3",
//		"mm":   "04",
//		"m":    "4",
//		"ss":   "05",
//		"s":    "5",
//		"a":    "PM",
//		"S":    ".0",
//		"SS":    ".00",
//		"SSS":   ".000",
//		"SSSN":  ".0000",
//		"SSSNN": ".00000",
//		"SSSNNN": ".000000",
//		"SSSNNNN": ".0000000",
//		"SSSNNNNN":".00000000",
//		"SSSNNNNNN":".000000000",
//		"z":    "MST",
//		"Z":    "-0700",
//		"X":    "-07",
//		"XX":    "-0700",
//		"XXX":  "-07:00",
//	}
//)

const JSISO = "2006-01-02T15:04:05.000Z07:00"
const ISO8601 = "2006-01-02T15:04:05"

func TimeToUnixMilli(time time.Time) int64 {
	return time.UnixNano() / 1e6
}

func InterfaceToUnixMilli(i interface{}, format string) (int64, error) {
	switch t := i.(type) {
	case int64:
		return t, nil
	case int:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case time.Time:
		return TimeToUnixMilli(t), nil
	case string:
		var ti time.Time
		var err error
		var f = JSISO
		if format != "" {
			f, err = convertFormat(format)
			if err != nil {
				return 0, err
			}
		}
		ti, err = time.Parse(f, t)
		if err != nil {
			return 0, err
		}
		return TimeToUnixMilli(ti), nil
	default:
		return 0, fmt.Errorf("unsupported type to convert to timestamp %v", t)
	}
}

func InterfaceToTime(i interface{}, format string) (time.Time, error) {
	switch t := i.(type) {
	case int64:
		return TimeFromUnixMilli(t), nil
	case int:
		return TimeFromUnixMilli(int64(t)), nil
	case float64:
		return TimeFromUnixMilli(int64(t)), nil
	case time.Time:
		return t, nil
	case string:
		var ti time.Time
		var err error
		var f = JSISO
		if format != "" {
			f, err = convertFormat(format)
			if err != nil {
				return ti, err
			}
		}
		ti, err = time.Parse(f, t)
		if err != nil {
			return ti, err
		}
		return ti, nil
	default:
		return time.Now(), fmt.Errorf("unsupported type to convert to timestamp %v", t)
	}
}

func TimeFromUnixMilli(t int64) time.Time {
	return time.Unix(t/1000, (t%1000)*1e6).UTC()
}

func ParseTime(t string, f string) (time.Time, error) {
	if f, err := convertFormat(f); err != nil {
		return time.Now(), err
	} else {
		return time.Parse(f, t)
	}
}

func FormatTime(time time.Time, f string) (string, error) {
	if f, err := convertFormat(f); err != nil {
		return "", err
	} else {
		return time.Format(f), nil
	}
}

//func convertFormat(f string) string {
//	re := regexp.MustCompile(`(?m)(M{4})|(M{3})|(M{2})|(M{1})|(Y{4})|(Y{2})|(y{4})|(y{2})|(G{1})|(E{4})|(E{3})|(d{2})|(d{1})|(H{2})|(h{2})|(h{1})|(m{2})|(m{1})|(s{2})|(s{1})|(a{1})|(S{3}N{6})|(S{3}N{5})|(S{3}N{4})|(S{3}N{3})|(S{3}N{2})|(S{3}N{1})|(S{3})|(S{2})|(S{1})|(z{1})|(Z{1})|(X{3})|(X{2})|(X{1})`)
//	for _, match := range re.FindAllString(f, -1) {
//		for key, val := range formats {
//			if match == key {
//				f = strings.Replace(f, match, val, -1)
//			}
//		}
//	}
//	return f
//}

func convertFormat(f string) (string, error) {
	formatRune := []rune(f)
	lenFormat := len(formatRune)
	out := ""
	for i := 0; i < len(formatRune); i++ {
		switch r := formatRune[i]; r {
		case 'Y', 'y':
			j := 1
			for ; i+j < lenFormat && j <= 4; j++ {
				if formatRune[i+j] != r {
					break
				}
			}
			i = i + j - 1
			switch j {
			case 4: // YYYY
				out += "2006"
			case 2: // YY
				out += "06"
			default:
				return "", fmt.Errorf("invalid time format %s for Y/y", f)
			}
		case 'G': //era
			out += "AD"
		case 'M': // M MM MMM MMMM month of year
			j := 1
			for ; i+j < lenFormat && j <= 4; j++ {
				if formatRune[i+j] != r {
					break
				}
			}
			i = i + j - 1
			switch j {
			case 1: // M
				out += "1"
			case 2: // MM
				out += "01"
			case 3: // MMM
				out += "Jan"
			case 4: // MMMM
				out += "January"
			}
		case 'd': // d dd day of month
			j := 1
			for ; i+j < lenFormat && j <= 2; j++ {
				if formatRune[i+j] != r {
					break
				}

			}
			i = i + j - 1
			switch j {
			case 1: // d
				out += "2"
			case 2: // dd
				out += "02"
			}
		case 'E': // M MM MMM MMMM month of year
			j := 1
			for ; i+j < lenFormat && j <= 4; j++ {
				if formatRune[i+j] != r {
					break
				}
			}
			i = i + j - 1
			switch j {
			case 3: // EEE
				out += "Mon"
			case 4: // EEEE
				out += "Monday"
			default:
				return "", fmt.Errorf("invalid time format %s for E", f)
			}
		case 'H': // HH
			j := 1
			for ; i+j < lenFormat && j <= 2; j++ {
				if formatRune[i+j] != r {
					break
				}

			}
			i = i + j - 1
			switch j {
			case 2: // HH
				out += "15"
			default:
				return "", fmt.Errorf("invalid time format %s of H, only HH is supported", f)
			}
		case 'h': // h hh
			j := 1
			for ; i+j < lenFormat && j <= 2; j++ {
				if formatRune[i+j] != r {
					break
				}
			}
			i = i + j - 1
			switch j {
			case 1: // h
				out += "3"
			case 2: // hh
				out += "03"
			}
		case 'a': // a
			out += "PM"
		case 'm': // m mm minute of hour
			j := 1
			for ; i+j < lenFormat && j <= 2; j++ {
				if formatRune[i+j] != r {
					break
				}

			}
			i = i + j - 1
			switch j {
			case 1: // m
				out += "4"
			case 2: // mm
				out += "04"
			}
		case 's': // s ss
			j := 1
			for ; i+j < lenFormat && j <= 2; j++ {
				if formatRune[i+j] != r {
					break
				}

			}
			i = i + j - 1
			switch j {
			case 1: // s
				out += "5"
			case 2: // ss
				out += "05"
			}

		case 'S': // S SS SSS
			j := 1
			for ; i+j < lenFormat && j <= 3; j++ {
				if formatRune[i+j] != r {
					break
				}
			}
			i = i + j - 1
			switch j {
			case 1: // S
				out += ".0"
			case 2: // SS
				out += ".00"
			case 3: // SSS
				out += ".000"
			}
		case 'z': // z
			out += "MST"
		case 'Z': // Z
			out += "-0700"
		case 'X': // X XX XXX
			j := 1
			for ; i+j < lenFormat && j <= 3; j++ {
				if formatRune[i+j] != r {
					break
				}
			}
			i = i + j - 1
			switch j {
			case 1: // X
				out += "-07"
			case 2: // XX
				out += "-0700"
			case 3: // XXX
				out += "-07:00"
			}
		case '\'': // ' (text delimiter)  or '' (real quote)

			// real quote
			if formatRune[i+1] == r {
				out += "'"
				i = i + 1
				continue
			}

			tmp := []rune{}
			j := 1
			for ; i+j < lenFormat; j++ {
				if formatRune[i+j] != r {
					tmp = append(tmp, formatRune[i+j])
					continue
				}
				break
			}
			i = i + j
			out += string(tmp)
		default:
			out += string(r)
		}
	}
	return out, nil
}

//Time related. For Mock
func GetTicker(duration int) *clock.Ticker {
	return Clock.Ticker(time.Duration(duration) * time.Millisecond)
}

func GetTimer(duration int) *clock.Timer {
	return Clock.Timer(time.Duration(duration) * time.Millisecond)
}

func GetNowInMilli() int64 {
	return TimeToUnixMilli(Clock.Now())
}
