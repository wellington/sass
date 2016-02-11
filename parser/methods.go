package parser

import "github.com/wellington/sass/builtin/colors"

func init() {
	Register("rgb($green:0, $red:0, $blue:0)", colors.RGB)
	Register("rgba($green:0, $red:0, $blue:0, $alpha:0)", colors.RGBA)
}
