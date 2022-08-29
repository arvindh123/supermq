package starter

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"time"
)

var (
	ErrStructValPrt = errors.New("expected struct pointer")
)

type EnvTypes interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string
}

func ConvertStringToType[T EnvTypes](valueStr string) T {
	var value T
	switch p := any(&value).(type) {
	case *int:
		i, _ := strconv.ParseInt(valueStr, 10, 0)
		*p = int(i)
		return value
	case *int8:
		i, _ := strconv.ParseInt(valueStr, 10, 8)
		*p = int8(i)
		return value
	case *int16:
		i, _ := strconv.ParseInt(valueStr, 10, 16)
		*p = int16(i)
		return value
	case *int32:
		i, _ := strconv.ParseInt(valueStr, 10, 32)
		*p = int32(i)
		return value
	case *int64:
		i, _ := strconv.ParseInt(valueStr, 10, 64)
		*p = i
		return value
	case *uint:
		i, _ := strconv.ParseUint(valueStr, 10, 0)
		*p = uint(i)
		return value
	case *uint8:
		i, _ := strconv.ParseUint(valueStr, 10, 8)
		*p = uint8(i)
		return value
	case *uint16:
		i, _ := strconv.ParseUint(valueStr, 10, 16)
		*p = uint16(i)
		return value
	case *uint32:
		i, _ := strconv.ParseUint(valueStr, 10, 32)
		*p = uint32(i)
		return value
	case *uint64:
		i, _ := strconv.ParseUint(valueStr, 10, 64)
		*p = i
		return value
	case *float32:
		i, _ := strconv.ParseFloat(valueStr, 32)
		*p = float32(i)
		return value
	case *float64:
		i, _ := strconv.ParseFloat(valueStr, 64)
		*p = i
		return value
	case *string:
		*p = valueStr
		return value
	default:
		return any(valueStr).(T)
	}
}

func GetEnv[T EnvTypes](key string, fallback T) T {
	envValue := os.Getenv(key)
	if envValue == "" {
		return fallback
	}
	var value T
	switch p := any(&value).(type) {
	case *int:
		i, _ := strconv.ParseInt(envValue, 10, 0)
		*p = int(i)
		return value
	case *int8:
		i, _ := strconv.ParseInt(envValue, 10, 8)
		*p = int8(i)
		return value
	case *int16:
		i, _ := strconv.ParseInt(envValue, 10, 16)
		*p = int16(i)
		return value
	case *int32:
		i, _ := strconv.ParseInt(envValue, 10, 32)
		*p = int32(i)
		return value
	case *int64:
		i, _ := strconv.ParseInt(envValue, 10, 64)
		*p = i
		return value
	case *uint:
		i, _ := strconv.ParseUint(envValue, 10, 0)
		*p = uint(i)
		return value
	case *uint8:
		i, _ := strconv.ParseUint(envValue, 10, 8)
		*p = uint8(i)
		return value
	case *uint16:
		i, _ := strconv.ParseUint(envValue, 10, 16)
		*p = uint16(i)
		return value
	case *uint32:
		i, _ := strconv.ParseUint(envValue, 10, 32)
		*p = uint32(i)
		return value
	case *uint64:
		i, _ := strconv.ParseUint(envValue, 10, 64)
		*p = i
		return value
	case *float32:
		i, _ := strconv.ParseFloat(envValue, 32)
		*p = float32(i)
		return value
	case *float64:
		i, _ := strconv.ParseFloat(envValue, 64)
		*p = i
		return value
	case *string:
		*p = envValue
		return value
	default:
		return fallback
	}
}

func LoadConfig(config interface{}) error {
	valPtr := reflect.ValueOf(config)

	if valPtr.Kind() != reflect.Ptr {
		return ErrStructValPrt
	}

	val := valPtr.Elem()

	if val.Kind() != reflect.Struct {
		return ErrStructValPrt
	}
	typ := val.Type()

	errStr := ""
	for i := 0; i < typ.NumField(); i++ {
		envVar := typ.Field(i).Tag.Get("env")
		defValStr := typ.Field(i).Tag.Get("default")
		name := typ.Field(i).Name

		switch val.FieldByName(name).Type() {
		case reflect.TypeOf(time.Nanosecond):
			setVal := GetEnv(envVar, ConvertStringToType[string](defValStr))
			setValDur, err := time.ParseDuration(setVal)
			if err != nil {
				errStr = fmt.Sprintf("%sfield name : %s , error: %s ;", errStr, name, err.Error())
			} else {
				val.FieldByName(name).SetInt(int64(setValDur))
			}
		default:
			switch typ.Field(i).Type.Kind() {

			case reflect.Int8:
				setVal := GetEnv(envVar, ConvertStringToType[int8](defValStr))
				val.FieldByName(name).SetInt(int64(setVal))
			case reflect.Int16:
				setVal := GetEnv(envVar, ConvertStringToType[int16](defValStr))
				val.FieldByName(name).SetInt(int64(setVal))
			case reflect.Int32:
				setVal := GetEnv(envVar, ConvertStringToType[int32](defValStr))
				val.FieldByName(name).SetInt(int64(setVal))
			case reflect.Int64:

			case reflect.Int:
				setVal := GetEnv(envVar, ConvertStringToType[int](defValStr))
				val.FieldByName(name).SetInt(int64(setVal))

			case reflect.Uint8:
				setVal := GetEnv(envVar, ConvertStringToType[uint8](defValStr))
				val.FieldByName(name).SetUint(uint64(setVal))
			case reflect.Uint16:
				setVal := GetEnv(envVar, ConvertStringToType[uint16](defValStr))
				val.FieldByName(name).SetUint(uint64(setVal))
			case reflect.Uint32:
				setVal := GetEnv(envVar, ConvertStringToType[uint32](defValStr))
				val.FieldByName(name).SetUint(uint64(setVal))
			case reflect.Uint64:
				setVal := GetEnv(envVar, ConvertStringToType[uint64](defValStr))
				val.FieldByName(name).SetUint(uint64(setVal))
			case reflect.Uint:
				setVal := GetEnv(envVar, ConvertStringToType[uint](defValStr))
				val.FieldByName(name).SetUint(uint64(setVal))

			case reflect.String:
				setVal := GetEnv(envVar, ConvertStringToType[string](defValStr))
				val.FieldByName(name).SetString(setVal)

			default:
				errStr = fmt.Sprintf("%sfield name : %s , datatype: %s datatype is not supported;", errStr, name, typ.Field(i).Type.Kind().String())
			}
		}
	}

	if errStr == "" {
		return nil
	}

	return errors.New(fmt.Sprintf("following fields and their respective error : %s", errStr))
}
