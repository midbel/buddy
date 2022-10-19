package builtins

var Builtins = map[string]func(...any) (any, error){
	"len":    Len,
	"upper":  Upper,
	"lower":  Lower,
	"printf": Printf,
	"print":  Print,
}
