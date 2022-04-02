package examplegolangdicontainer

import (
	"reflect"
)

type (
	container struct {
		factoryInfos                map[reflect.Type]factoryInfo
		cache                       map[reflect.Type]reflect.Value
		containerInterfaceType      reflect.Type
		ioCContainerInterfaceType   reflect.Type
		serviceLocatorInterfaceType reflect.Type
	}
	Container interface {
		Register(constructor Target) error
		IoCContainer
	}
	IoCContainer interface {
		ServiceLocator
	}
	ServiceLocator interface {
		Invoke(invoker Invoker) error
	}
	ContainerOptions struct{}
	factoryInfo struct {
		target        reflect.Value
		ins           []reflect.Type
		isFunc        bool
		lifetimeScope LifetimeScope
	}
	Invoker interface{}
	 LifetimeScope int
	Target interface{}

)
const (
	ContainerManaged LifetimeScope = iota
	InvokeManaged
)

func NewContainer() Container {
	return &container{
		factoryInfos:                make(map[reflect.Type]factoryInfo),
		cache:                       make(map[reflect.Type]reflect.Value),
		containerInterfaceType:      reflect.TypeOf((*Container)(nil)).Elem(),
		ioCContainerInterfaceType:   reflect.TypeOf((*IoCContainer)(nil)).Elem(),
		serviceLocatorInterfaceType: reflect.TypeOf((*ServiceLocator)(nil)).Elem(),
	}
}

func (c *container) Register(target Target) error {
	funcReturn, funcParams, err := getTargetInitFuncInfos(target)
	if err != nil {
		return err
	}
	lts := InvokeManaged
	kind := funcReturn.Kind()
	isFunc := funcParams != nil
	if !isFunc {
		lts = ContainerManaged
	}
	value := reflect.ValueOf(target)
	if kind != reflect.Ptr {
		c.factoryInfos[funcReturn] = factoryInfo{target: value, lifetimeScope: lts, ins: funcParams, isFunc: isFunc}
		_, ok := c.cache[funcReturn]
		if ok {
			delete(c.cache, funcReturn)
		}
		return nil
	} 
	return ErrUnableRegistFunction
}

func (c *container) Invoke(invoker Invoker) error {
	t := reflect.TypeOf(invoker)
	if t.Kind() != reflect.Func {
		return ErrRequireFunction
	}
	params := getFuncParams(t)
	lenParams := len(params)
	if lenParams == 0 {
		return ErrNotFoundComponent
	}
	args := make([]reflect.Value, lenParams)
	cache := make(map[reflect.Type]reflect.Value)
	for i, funcParam := range params {
		// function の型から対象の function param を取得して、順番に配列に詰める
		p, err := c.resolve(funcParam, &cache)
		if err != nil {
			return err
		}
		args[i] = *p
	}

	fn := reflect.ValueOf(invoker)
	outs := fn.Call(args)
	if err := c.getError(outs); err != nil {
		return err
	}
	return nil
}

func (c *container) resolve(t reflect.Type, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	if c.containerInterfaceType == t || c.ioCContainerInterfaceType == t || c.serviceLocatorInterfaceType == t {
		v := reflect.ValueOf(c)
		return &v, nil
	}
	// type から対象の function を取得するための振り分け
	factoryInfo, ok := c.factoryInfos[t]
	if !ok {
		return nil, newErrInvalidResolveComponent(t)
	}
	switch factoryInfo.lifetimeScope {
	case ContainerManaged:
		return c.resolveContainerManagedObject(t, factoryInfo, cache)
	}
	return c.resolveInvokeManagedObject(t, factoryInfo, cache)
}

func (c *container) resolveContainerManagedObject(t reflect.Type, factoryInfo factoryInfo, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	// MEMO: 対象 function の param を再起的に取得し、対象 functions を call する
	if v, ok := c.cache[t]; ok {
		return &v, nil
	}
	lenIns := len(factoryInfo.ins)
	args := make([]reflect.Value, lenIns)
	for i, in := range factoryInfo.ins {
		v, err := c.resolve(in, cache)
		if err != nil {
			return nil, err
		}
		args[i] = *v
	}

	outs := factoryInfo.target.Call(args)
	for _, out := range outs {
		if err, ok := out.Interface().(error); ok {
			return nil, err
		}
	}
	if err := c.getError(outs); err != nil {
		return nil, err
	}
	out := outs[0]
	c.cache[t] = out
	return &out, nil
}

func (c *container) resolveInvokeManagedObject(t reflect.Type, factoryInfo factoryInfo, cache *map[reflect.Type]reflect.Value) (*reflect.Value, error) {
	// MEMO: 対象 function の param を再起的に取得し、対象 functions を call する
	cch := *cache
	if v, ok := cch[t]; ok {
		return &v, nil
	}
	if !factoryInfo.isFunc {
		cch[t] = factoryInfo.target
		return &factoryInfo.target, nil
	}
	lenIns := len(factoryInfo.ins)
	args := make([]reflect.Value, lenIns)
	for i, in := range factoryInfo.ins {
		v, err := c.resolve(in, cache)
		if err != nil {
			return nil, err
		}
		args[i] = *v
	}

	outs := factoryInfo.target.Call(args)
	if err := c.getError(outs); err != nil {
		return nil, err
	}
	out := outs[0]
	cch[t] = out
	return &out, nil
}
