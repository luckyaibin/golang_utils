//decorataor or interceptor in go
package main

import (
	"fmt"
	"reflect"
)
//Decorator in go
func Decorator(decoPtr,fn interface{}) (err error){
	var decoratedFunc reflect.Value
	var targetFunc	reflect.Value

	decoratedFunc = reflect.ValueOf(decoPtr).Elem()
	targetFunc = reflect.ValueOf(fn)

	inner := func(in []reflect.Value)(out []reflect.Value){
		fmt.Println("Decorator before",in)
		out = targetFunc.Call(in)
		fmt.Println("Decorator after",out)
		return
	}
	wrapper := reflect.MakeFunc(targetFunc.Type(),inner)
	decoratedFunc.Set(wrapper)
	return
}

//Interceptor in go.
//添加到前后的拦截器必须和fn有相同的参数签名：因为preInterFunc.Call(in)，afterInterFunc.Call(in)这样调用，传入的参数
func Interceptor(decoPtr,fn interface{},preInter interface{},afterInter interface{}) (err error){
	var decoratedFunc reflect.Value
	var targetFunc	reflect.Value

	decoratedFunc = reflect.ValueOf(decoPtr).Elem()
	targetFunc = reflect.ValueOf(fn)

	inner := func(in []reflect.Value)(out []reflect.Value){
		if nil != preInter {
			preInterFunc := reflect.ValueOf(preInter)
			if !preInterFunc.IsNil(){
				//fmt.Println("Decorator before",in)
				preInterFunc.Call(in)
			}
		}
		out = targetFunc.Call(in)

		if nil != afterInter {
			afterInterFunc := reflect.ValueOf(afterInter)
			if !afterInterFunc.IsNil() {
				//fmt.Println("Decorator after",out)
				afterInterFunc.Call(in)
			}
		}
		return
	}
	wrapper := reflect.MakeFunc(targetFunc.Type(),inner)
	decoratedFunc.Set(wrapper)
	return
}


func add(a,b *int) (int,int,string){
	fmt.Println("in add function")
	return *a+*b,-100,"Im return string."
}
//拦截器输入参数必须和被拦截的一样，返回值则无所谓
func preInter(a,b *int) int{
	fmt.Println("I am preInter")
	*a = 10000
	return *a-*b
}
func afterInter(a ,b *int) (int ,string){
	fmt.Println("I after preInter")
	*b = 20000
	return *a* *b ,"Hi"
}
 
func main(){
	addtype:=add
  Interceptor(&addtype,add,preInter,afterInter)
	Interceptor(&addtype,addtype,nil,afterInter)
	fmt.Println(addtype(&a,&b))
}
