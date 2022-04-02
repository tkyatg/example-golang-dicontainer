package examplegolangdicontainer

import (
	"fmt"
	"reflect"
	"strings"
)

var (
	ErrUnableRegistFunction = fmt.Errorf("登録できない Function です")
	ErrRequireFunction                   = fmt.Errorf("関数を指定してください")
	ErrNotFoundComponent                 = fmt.Errorf("解決するオブジェクトが存在しません")
	ErrRequireResponse                   = fmt.Errorf("登録する関数には返り値が必要です")
)

func newErrInvalidResolveComponent(t reflect.Type) error {
	return fmt.Errorf("指定されたタイプを解決できません。(%v)", t)
}
func IsErrInvalidResolveComponent(err error) bool {
	return strings.HasPrefix(err.Error(), "指定されたタイプを解決できません。")
}
