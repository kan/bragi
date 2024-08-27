package dict

import (
	"log"
	"time"
)

type LispDict struct {
	DateFormat     string
	DateTimeFormat string
	Location       *time.Location
}

func (d *LispDict) Convert(word string) ([]string, error) {
	now := time.Now().In(d.Location)

	if word == "きょう" || word == "ほんじつ" || word == "today" {
		return []string{now.Format("2006-01-02")}, nil
	} else if word == "きのう" || word == "さくじつ" || word == "yesterday" {
		return []string{now.AddDate(0, 0, -1).Format("2006-01-02")}, nil
	} else if word == "あす" || word == "よくじつ" || word == "tomorrow" {
		return []string{now.AddDate(0, 0, 1).Format("2006-01-02")}, nil
	} else if word == "いま" || word == "げんざい" || word == "now" {
		return []string{now.AddDate(0, 0, 1).Format("2006-01-02 15:04")}, nil
	}
	return []string{}, nil
}

func NewLispDict(df, dtf, tz string) *LispDict {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Println(err)
		loc = time.Local
	}

	return &LispDict{
		DateFormat:     df,
		DateTimeFormat: df,
		Location:       loc,
	}
}
