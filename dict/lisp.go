package dict

import "time"

type LispDict struct{}

func (d *LispDict) Convert(word string) ([]string, error) {
	if word == "きょう" || word == "ほんじつ" || word == "today" {
		return []string{time.Now().Format("2006-01-02")}, nil
	} else if word == "きのう" || word == "さくじつ" || word == "yesterday" {
		return []string{time.Now().AddDate(0, 0, -1).Format("2006-01-02")}, nil
	} else if word == "あす" || word == "よくじつ" || word == "tomorrow" {
		return []string{time.Now().AddDate(0, 0, 1).Format("2006-01-02")}, nil
	} else if word == "いま" || word == "げんざい" || word == "now" {
		return []string{time.Now().AddDate(0, 0, 1).Format("2006-01-02 15:04")}, nil
	}
	return []string{}, nil
}

func NewLispDict() *LispDict {
	return &LispDict{}
}
