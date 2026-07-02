package theme

import (
	"maps"
	"strings"
)

type Palette struct {
	Title           string            `json:"title"`
	Number          string            `json:"number"`
	Hint            string            `json:"hint"`
	Flag            string            `json:"flag"`
	BadgeForeground string            `json:"badge_foreground"`
	BadgeBackground string            `json:"badge_background"`
	Prompt          string            `json:"prompt"`
	OK              string            `json:"ok"`
	Skip            string            `json:"skip"`
	Fail            string            `json:"fail"`
	Dim             string            `json:"dim"`
	Selected        string            `json:"selected"`
	Error           string            `json:"error"`
	Unavailable     string            `json:"unavailable"`
	Categories      map[string]string `json:"categories"`
}

func Default() Palette {
	return Nord()
}

func Classic() Palette {
	return Palette{
		Title:           "205",
		Number:          "39",
		Hint:            "245",
		Flag:            "39",
		BadgeForeground: "230",
		BadgeBackground: "62",
		Prompt:          "86",
		OK:              "42",
		Skip:            "214",
		Fail:            "196",
		Dim:             "245",
		Selected:        "42",
		Error:           "196",
		Unavailable:     "245",
		Categories: map[string]string{
			"directory":   "111",
			"image":       "205",
			"animation":   "207",
			"video":       "170",
			"audio":       "141",
			"archive":     "214",
			"font":        "45",
			"disk":        "196",
			"vm":          "202",
			"doc":         "42",
			"ebook":       "36",
			"spreadsheet": "76",
			"data":        "39",
			"geo":         "70",
			"schema":      "99",
			"diagram":     "135",
			"code":        "48",
			"custom":      "245",
		},
	}
}

func Nord() Palette {
	return Palette{
		Title:           "#B48EAD",
		Number:          "#88C0D0",
		Hint:            "#D8DEE9",
		Flag:            "#88C0D0",
		BadgeForeground: "#2E3440",
		BadgeBackground: "#81A1C1",
		Prompt:          "#8FBCBB",
		OK:              "#A3BE8C",
		Skip:            "#EBCB8B",
		Fail:            "#BF616A",
		Dim:             "#616E88",
		Selected:        "#A3BE8C",
		Error:           "#BF616A",
		Unavailable:     "#616E88",
		Categories: map[string]string{
			"directory":   "#88C0D0",
			"image":       "#B48EAD",
			"animation":   "#D08770",
			"video":       "#5E81AC",
			"audio":       "#81A1C1",
			"archive":     "#EBCB8B",
			"font":        "#8FBCBB",
			"disk":        "#BF616A",
			"vm":          "#D08770",
			"doc":         "#A3BE8C",
			"ebook":       "#8FBCBB",
			"spreadsheet": "#A3BE8C",
			"data":        "#88C0D0",
			"geo":         "#A3BE8C",
			"schema":      "#B48EAD",
			"diagram":     "#D08770",
			"code":        "#8FBCBB",
			"custom":      "#D8DEE9",
		},
	}
}

func Named(name string) Palette {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "", "default", "nord":
		return Nord()
	case "classic", "legacy":
		return Classic()
	default:
		return Default()
	}
}

func (p Palette) Merge(overrides Palette) Palette {
	result := p
	result.Categories = make(map[string]string, len(p.Categories))
	maps.Copy(result.Categories, p.Categories)
	mergeString := func(target *string, value string) {
		if strings.TrimSpace(value) != "" {
			*target = strings.TrimSpace(value)
		}
	}
	mergeString(&result.Title, overrides.Title)
	mergeString(&result.Number, overrides.Number)
	mergeString(&result.Hint, overrides.Hint)
	mergeString(&result.Flag, overrides.Flag)
	mergeString(&result.BadgeForeground, overrides.BadgeForeground)
	mergeString(&result.BadgeBackground, overrides.BadgeBackground)
	mergeString(&result.Prompt, overrides.Prompt)
	mergeString(&result.OK, overrides.OK)
	mergeString(&result.Skip, overrides.Skip)
	mergeString(&result.Fail, overrides.Fail)
	mergeString(&result.Dim, overrides.Dim)
	mergeString(&result.Selected, overrides.Selected)
	mergeString(&result.Error, overrides.Error)
	mergeString(&result.Unavailable, overrides.Unavailable)
	for key, value := range overrides.Categories {
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		if key != "" && value != "" {
			result.Categories[key] = value
		}
	}
	return result
}

func (p Palette) CategoryColor(category string) string {
	if p.Categories != nil {
		if color := strings.TrimSpace(p.Categories[strings.ToLower(strings.TrimSpace(category))]); color != "" {
			return color
		}
		if color := strings.TrimSpace(p.Categories["custom"]); color != "" {
			return color
		}
	}
	return "245"
}
