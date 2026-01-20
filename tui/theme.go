package tui

import "github.com/gdamore/tcell/v3"

type Theme struct {
	Name     string
	bg       tcell.Color
	fg       tcell.Color
	red      tcell.Color
	green    tcell.Color
	yellow   tcell.Color
	blue     tcell.Color
	purple   tcell.Color
	aqua     tcell.Color
	orange   tcell.Color
	gray     tcell.Color
	darkGray tcell.Color
	headerBg tcell.Color
	headerFg tcell.Color
	footerBg tcell.Color
	footerFg tcell.Color
	sizeFg   tcell.Color
	buttonBg tcell.Color
	buttonFg tcell.Color
	modalBg  tcell.Color
	modalFg  tcell.Color
}

var themes = map[string]Theme{
	"gruvbox-dark": {
		Name:     "Gruvbox Dark",
		bg:       tcell.NewRGBColor(40, 40, 40),
		fg:       tcell.NewRGBColor(235, 219, 178),
		red:      tcell.NewRGBColor(204, 36, 29),
		green:    tcell.NewRGBColor(152, 151, 26),
		yellow:   tcell.NewRGBColor(215, 153, 33),
		blue:     tcell.NewRGBColor(69, 133, 136),
		purple:   tcell.NewRGBColor(177, 98, 134),
		aqua:     tcell.NewRGBColor(104, 157, 106),
		orange:   tcell.NewRGBColor(214, 93, 14),
		gray:     tcell.NewRGBColor(146, 131, 116),
		darkGray: tcell.NewRGBColor(60, 56, 54),
		headerBg: tcell.NewRGBColor(214, 93, 14),
		headerFg: tcell.NewRGBColor(60, 56, 54),
		footerBg: tcell.NewRGBColor(60, 56, 54),
		footerFg: tcell.NewRGBColor(235, 219, 178),
		sizeFg:   tcell.NewRGBColor(215, 153, 33),
		buttonBg: tcell.NewRGBColor(214, 93, 14),
		buttonFg: tcell.NewRGBColor(60, 56, 54),
		modalBg:  tcell.NewRGBColor(40, 40, 40),
		modalFg:  tcell.NewRGBColor(235, 219, 178),
	},
	"nord": {
		Name:     "Nord",
		bg:       tcell.NewRGBColor(46, 52, 64),
		fg:       tcell.NewRGBColor(216, 222, 233),
		red:      tcell.NewRGBColor(191, 97, 106),
		green:    tcell.NewRGBColor(163, 190, 140),
		yellow:   tcell.NewRGBColor(235, 203, 139),
		blue:     tcell.NewRGBColor(136, 192, 208),
		purple:   tcell.NewRGBColor(129, 161, 193),
		aqua:     tcell.NewRGBColor(143, 188, 187),
		orange:   tcell.NewRGBColor(191, 97, 106),
		gray:     tcell.NewRGBColor(59, 66, 82),
		darkGray: tcell.NewRGBColor(67, 76, 94),
		headerBg: tcell.NewRGBColor(129, 161, 193),
		headerFg: tcell.NewRGBColor(46, 52, 64),
		footerBg: tcell.NewRGBColor(67, 76, 94),
		footerFg: tcell.NewRGBColor(216, 222, 233),
		sizeFg:   tcell.NewRGBColor(235, 203, 139),
		buttonBg: tcell.NewRGBColor(129, 161, 193),
		buttonFg: tcell.NewRGBColor(46, 52, 64),
		modalBg:  tcell.NewRGBColor(46, 52, 64),
		modalFg:  tcell.NewRGBColor(216, 222, 233),
	},
	"catppuccin": {
		Name:     "Catppuccin Mocha",
		bg:       tcell.NewRGBColor(30, 30, 46),
		fg:       tcell.NewRGBColor(205, 214, 244),
		red:      tcell.NewRGBColor(243, 139, 168),
		green:    tcell.NewRGBColor(166, 227, 161),
		yellow:   tcell.NewRGBColor(249, 226, 175),
		blue:     tcell.NewRGBColor(137, 180, 250),
		purple:   tcell.NewRGBColor(203, 166, 247),
		aqua:     tcell.NewRGBColor(148, 226, 213),
		orange:   tcell.NewRGBColor(250, 179, 135),
		gray:     tcell.NewRGBColor(110, 109, 128),
		darkGray: tcell.NewRGBColor(49, 50, 68),
		headerBg: tcell.NewRGBColor(137, 180, 250),
		headerFg: tcell.NewRGBColor(30, 30, 46),
		footerBg: tcell.NewRGBColor(49, 50, 68),
		footerFg: tcell.NewRGBColor(205, 214, 244),
		sizeFg:   tcell.NewRGBColor(249, 226, 175),
		buttonBg: tcell.NewRGBColor(137, 180, 250),
		buttonFg: tcell.NewRGBColor(30, 30, 46),
		modalBg:  tcell.NewRGBColor(30, 30, 46),
		modalFg:  tcell.NewRGBColor(205, 214, 244),
	},
	"dracula": {
		Name:     "Dracula",
		bg:       tcell.NewRGBColor(40, 42, 54),
		fg:       tcell.NewRGBColor(248, 248, 242),
		red:      tcell.NewRGBColor(255, 85, 85),
		green:    tcell.NewRGBColor(80, 250, 123),
		yellow:   tcell.NewRGBColor(241, 250, 140),
		blue:     tcell.NewRGBColor(139, 233, 253),
		purple:   tcell.NewRGBColor(189, 147, 249),
		aqua:     tcell.NewRGBColor(139, 233, 253),
		orange:   tcell.NewRGBColor(255, 184, 108),
		gray:     tcell.NewRGBColor(68, 71, 90),
		darkGray: tcell.NewRGBColor(68, 71, 90),
		headerBg: tcell.NewRGBColor(189, 147, 249),
		headerFg: tcell.NewRGBColor(40, 42, 54),
		footerBg: tcell.NewRGBColor(68, 71, 90),
		footerFg: tcell.NewRGBColor(248, 248, 242),
		sizeFg:   tcell.NewRGBColor(255, 184, 108),
		buttonBg: tcell.NewRGBColor(189, 147, 249),
		buttonFg: tcell.NewRGBColor(40, 42, 54),
		modalBg:  tcell.NewRGBColor(40, 42, 54),
		modalFg:  tcell.NewRGBColor(248, 248, 242),
	},
}

func getThemeNames() []string {
	var names []string
	for n := range themes {
		names = append(names, n)
	}
	return names
}
