package pkg11

import (
	pkg2notsamebasedir "www.example.com/hoge/fuga/pkg2"
	"www.example.com/hoge/fuga/pkg3"
)

var A string = pkg3.A
var B string = pkg2notsamebasedir.A
