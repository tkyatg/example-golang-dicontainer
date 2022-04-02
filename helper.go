package examplegolangdicontainer

import "reflect"

// MEMO: function の 引数の型を返す
func getFuncParams(t reflect.Type) []reflect.Type {
	len := t.NumIn()
	in := make([]reflect.Type, len)
	for i := 0; i < len; i++ {
		in[i] = t.In(i)
	}
	return in
}

// MEMO: new func の戻り値の型返す
func getFuncOutPutFirstRes(t reflect.Type) (reflect.Type, error) {
	l := t.NumOut()
	if l < 1 {
		return nil, ErrRequireResponse
	}
	return t.Out(0), nil
}

// MEMO: 対象がfuncの場合、
func getTargetInitFuncInfos(target Target) (funcReturn  reflect.Type, funcParams []reflect.Type, err error) {
	t := reflect.TypeOf(target)
	if t.Kind() == reflect.Func {
		out, err := getFuncOutPutFirstRes(t)
		if err != nil {
			return nil, nil, err
		}
		ins := getFuncParams(t)
		return out, ins, nil
	}
	return t, nil, nil
}

// MEMO: 引数から error が取得できた場合、その error を返す
func (c *container) getError(outs []reflect.Value) error {
	l := len(outs)
	if l > 0 {
		if err, ok := outs[l-1].Interface().(error); ok {
			return err
		}
	}
	return nil
}