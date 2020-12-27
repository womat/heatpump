package config

import "flag"

type flagStruct struct {
	flagType     int
	value        interface{}
	defaultValue interface{}
	usage        string
}

type flagm map[string]flagStruct

const (
	flagInt = iota
	flagBool
	flagString
)

func parse(flags flagm) {
	for name, f := range flags {
		switch f.flagType {
		case flagInt:
			f.value = flag.Int(name, f.defaultValue.(int), f.usage)
		case flagBool:
			f.value = flag.Bool(name, f.defaultValue.(bool), f.usage)
		case flagString:
			f.value = flag.String(name, f.defaultValue.(string), f.usage)
		}
		flags[name] = f
	}

	flag.Parse()
}

func (f flagm) bool(c string) bool {
	if f, ok := f[c]; ok {
		return castToBool(f.value)
	}
	return false
}

func (f flagm) string(c string) string {
	if f, ok := f[c]; ok {
		return castToString(f.value)
	}

	return ""
}

func castToString(c interface{}) string {
	switch v := c.(type) {
	case *string:
		return *v
	}
	return ""
}

func castToBool(c interface{}) bool {
	switch v := c.(type) {
	case *bool:
		return *v
	}
	return false
}
