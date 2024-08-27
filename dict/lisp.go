package dict

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type LispDict struct {
	YearFormat     string
	MonthFormat    string
	DateFormat     string
	DateTimeFormat string
	Location       *time.Location
}

var reNYearBefore *regexp.Regexp
var reNYearAfter *regexp.Regexp
var reNMonthBefore *regexp.Regexp
var reNMonthAfter *regexp.Regexp
var reNDayBefore *regexp.Regexp
var reNDayAfter *regexp.Regexp
var eraStartYears = map[string]int{
	"めいじ":   1868, // 明治
	"たいしょう": 1912, // 大正
	"しょうわ":  1926, // 昭和
	"へいせい":  1989, // 平成
	"れいわ":   2019, // 令和
}
var reEras map[string]*regexp.Regexp

func init() {
	reNYearBefore = regexp.MustCompile(`(\d+)ねんまえ`)
	reNYearAfter = regexp.MustCompile(`(\d+)ねんご`)
	reNMonthBefore = regexp.MustCompile(`(\d+)かげつまえ`)
	reNMonthAfter = regexp.MustCompile(`(\d+)かげつご`)
	reNDayBefore = regexp.MustCompile(`(\d+)にちまえ`)
	reNDayAfter = regexp.MustCompile(`(\d+)にちご`)
	reEras = map[string]*regexp.Regexp{}
	for era := range eraStartYears {
		reEras[era] = regexp.MustCompile(era + `(\d+)ねん`)
	}
}

func (d *LispDict) Convert(word string) ([]string, error) {
	now := time.Now().In(d.Location)

	if word == "ことし" || word == "ほんねん" {
		return []string{now.Format(d.YearFormat)}, nil
	} else if word == "さくねん" {
		return []string{now.AddDate(-1, 0, 0).Format(d.YearFormat)}, nil
	} else if word == "らいねん" {
		return []string{now.AddDate(1, 0, 0).Format(d.YearFormat)}, nil
	} else if word == "こんげつ" {
		return []string{now.Format(d.MonthFormat)}, nil
	} else if word == "せんげつ" || word == "ぜんげつ" {
		return []string{now.AddDate(0, -1, 0).Format(d.MonthFormat)}, nil
	} else if word == "らいげつ" || word == "よくげつ" {
		return []string{now.AddDate(0, 1, 0).Format(d.MonthFormat)}, nil
	} else if word == "きょう" || word == "ほんじつ" || word == "today" {
		return []string{now.Format(d.DateFormat)}, nil
	} else if word == "きのう" || word == "さくじつ" || word == "yesterday" {
		return []string{now.AddDate(0, 0, -1).Format(d.DateFormat)}, nil
	} else if word == "あす" || word == "よくじつ" || word == "tomorrow" {
		return []string{now.AddDate(0, 0, 1).Format(d.DateFormat)}, nil
	} else if word == "いま" || word == "げんざい" || word == "now" {
		return []string{now.AddDate(0, 0, 1).Format(d.DateTimeFormat)}, nil
	} else if ms := reNYearBefore.FindStringSubmatch(word); len(ms) > 1 {
		dy, err := strconv.Atoi(ms[1])
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		return []string{now.AddDate(-dy, 0, 0).Format(d.MonthFormat)}, nil
	} else if ms := reNYearAfter.FindStringSubmatch(word); len(ms) > 1 {
		dy, err := strconv.Atoi(ms[1])
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		return []string{now.AddDate(dy, 0, 0).Format(d.MonthFormat)}, nil
	} else if ms := reNMonthBefore.FindStringSubmatch(word); len(ms) > 1 {
		dm, err := strconv.Atoi(ms[1])
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		return []string{now.AddDate(0, -dm, 0).Format(d.MonthFormat)}, nil
	} else if ms := reNMonthAfter.FindStringSubmatch(word); len(ms) > 1 {
		dm, err := strconv.Atoi(ms[1])
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		return []string{now.AddDate(0, dm, 0).Format(d.MonthFormat)}, nil
	} else if ms := reNDayBefore.FindStringSubmatch(word); len(ms) > 1 {
		dd, err := strconv.Atoi(ms[1])
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		return []string{now.AddDate(0, 0, -dd).Format(d.MonthFormat)}, nil
	} else if ms := reNDayAfter.FindStringSubmatch(word); len(ms) > 1 {
		dd, err := strconv.Atoi(ms[1])
		if err != nil {
			return []string{}, errors.WithStack(err)
		}
		return []string{now.AddDate(0, 0, dd).Format(d.MonthFormat)}, nil
	}
	for era, re := range reEras {
		if ms := re.FindStringSubmatch(word); len(ms) > 1 {
			ey, err := strconv.Atoi(ms[1])
			if err != nil {
				return []string{}, errors.WithStack(err)
			}
			return []string{fmt.Sprintf("西暦%d年", eraStartYears[era]+ey-1)}, nil
		}
	}
	return []string{}, nil
}

func NewLispDict(yf, mf, df, dtf, tz string) *LispDict {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Println(err)
		loc = time.Local
	}

	return &LispDict{
		YearFormat:     yf,
		MonthFormat:    mf,
		DateFormat:     df,
		DateTimeFormat: df,
		Location:       loc,
	}
}
