package main

type MenuOpp struct{
	option string
	info string
}

func newOpp(option string, info string) MenuOpp{
	return MenuOpp{option: option, info: info}
}