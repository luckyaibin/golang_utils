 package main

import (
    "fmt"
    "reflect"
    "strings"
)

//type:interface value:sturct
func PrintStruct(t reflect.Type, v reflect.Value, pc int) {
    fmt.Println("")
    for i := 0; i < t.NumField(); i++ {
        var indent = strings.Repeat(" ", pc)
        var name = t.Field(i).Name                           //private value will panic reflect
        if !('A' <= name[0] && name[0] <= 'Z') {                 fmt.Print(indent, name, ":  *private*")
            fmt.Println("")
            continue
        }                                                    fmt.Print(indent, name, ":")                         value := v.Field(i)
        PrintVar(value.Interface(), pc+2)
        fmt.Println("")
    }
}

func PrintArraySlice(v reflect.Value, pc int) {
    for j := 0; j < v.Len(); j++ {
        PrintVar(v.Index(j).Interface(), pc+2)
    }
}
func PrintMap(v reflect.Value, pc int) {
    for _, k := range v.MapKeys() {
        PrintVar(k.Interface(), pc)
        PrintVar(v.MapIndex(k).Interface(), pc)
    }
}

func PrintVar(i interface{}, ident int) {
    t := reflect.TypeOf(i)
    v := reflect.ValueOf(i)
    if v.Kind() == reflect.Ptr {
        //nil ptr will crash
        if v.IsNil() {
            fmt.Print(strings.Repeat(" ",
                ident), "*nil* of ", t)
            return
        }                                                    // 如果v是nil,v.Type这会crash,                       //所以上面提前判断并返回
        v = reflect.ValueOf(i).Elem()
        t = v.Type()
    }                                                    switch v.Kind() {                                    case reflect.Array:
        PrintArraySlice(v, ident)
    case reflect.Chan:
        fmt.Print("Chan")
    case reflect.Func:
        fmt.Print("Func")
    case reflect.Interface:
        fmt.Print("Interface")
    case reflect.Map:
        PrintMap(v, ident)
    case reflect.Slice:
        PrintArraySlice(v, ident)
    case reflect.Struct:
        PrintStruct(t, v, ident)
    case reflect.UnsafePointer:
        fmt.Print("UnsafePointer")
    default:
        fmt.Print(strings.Repeat(" ", ident), v.Interface())
    }
}

var PPrint = PrintVar
