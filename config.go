package gson

import "bytes"
import "fmt"
import "encoding/json"

// NumberKind how to treat numbers.
type NumberKind byte

const (
	// FloatNumber to treat number as float64.
	FloatNumber NumberKind = iota + 1

	// SmartNumber to treat number as either integer or fall back to float64.
	SmartNumber
)

// MaxKeys maximum number of keys allowed in a property object. Affects
// memory pool. Changing this value will affect all new configuration objects.
var MaxKeys = 1024

// Config is the root object to access all transformations and APIs
// exported by this package. Before calling any of the config-methods,
// make sure to initialize them with desired settings.
//
// NOTE: Config objects are immutable.
type Config struct {
	nk    NumberKind
	pools mempools

	cborConfig
	jsonConfig
	collateConfig
	jptrConfig
	memConfig
}

// NewDefaultConfig return a new configuration with default settings:
//		+FloatNumber        +Stream
//		+UnicodeSpace       -strict
//		+doMissing          -arrayLenPrefix     +propertyLenPrefix
//		MaxJsonpointerLen
//		MaxStringLen        MaxKeys
//		MaxCollateLen       MaxJsonpointerLen
func NewDefaultConfig() *Config {
	config := &Config{
		nk: FloatNumber,
		cborConfig: cborConfig{
			ct: Stream,
		},
		jsonConfig: jsonConfig{
			ws:     UnicodeSpace,
			strict: false,
		},
		collateConfig: collateConfig{
			doMissing:         true,
			arrayLenPrefix:    false,
			propertyLenPrefix: true,
		},
		memConfig: memConfig{
			strlen:  MaxStringLen,
			numkeys: MaxKeys,
			itemlen: MaxCollateLen,
			ptrlen:  MaxJsonpointerLen,
		},
	}
	config = config.SetJptrlen(MaxJsonpointerLen)
	return config.init()
}

func (config *Config) init() *Config {
	config.buf = bytes.NewBuffer(make([]byte, 0, 1024)) // start with 1K
	config.enc = json.NewEncoder(config.buf)
	a, b, c, d := config.strlen, config.numkeys, config.itemlen, config.ptrlen
	config.pools = newMempool(a, b, c, d)
	return config
}

// SetNumberKind configure to interpret number values.
func (config Config) SetNumberKind(nk NumberKind) *Config {
	config.nk = nk
	return &config
}

// SetContainerEncoding configure to encode / decode cbor
// arrays and maps.
func (config Config) SetContainerEncoding(ct ContainerEncoding) *Config {
	config.ct = ct
	return &config
}

// SetSpaceKind setting to interpret whitespaces in json text.
func (config Config) SetSpaceKind(ws SpaceKind) *Config {
	config.ws = ws
	return &config
}

// SetStrict setting to enforce strict transforms to and from JSON.
// If set to true,
//   a. IntNumber configuration float numbers in JSON text still are parsed.
//   b. Use golang stdlib encoding/json for transforming strings to JSON.
func (config Config) SetStrict(what bool) *Config {
	config.strict = what
	return &config
}

// SortbyArrayLen setting to sort array of smaller-size before larger ones.
func (config Config) SortbyArrayLen(what bool) *Config {
	config.arrayLenPrefix = what
	return &config
}

// SortbyPropertyLen setting to sort properties of smaller size before
// larger ones.
func (config Config) SortbyPropertyLen(what bool) *Config {
	config.propertyLenPrefix = what
	return &config
}

// UseMissing setting to use TypeMissing collation.
func (config Config) UseMissing(what bool) *Config {
	config.doMissing = what
	return &config
}

// SetMaxkeys configure to set the maximum number of keys
// allowed in property item.
func (config Config) SetMaxkeys(n int) *Config {
	config.numkeys = n
	return config.init()
}

// ResetPools configure a new set of pools with specified size.
//	 strlen  - maximum length of string value inside JSON document
//	 numkeys - maximum number of keys that a property object can have
//	 itemlen - maximum length of collated value.
//	 ptrlen  - maximum possible length of json-pointer.
func (config Config) ResetPools(strlen, numkeys, itemlen, ptrlen int) *Config {
	config.memConfig = memConfig{
		strlen: strlen, numkeys: numkeys, itemlen: itemlen, ptrlen: ptrlen,
	}
	return config.init()
}

// NewCbor factory to create a new Cbor instance. Buffer can't be nil.
// If length is less than 0, ln will be assumed as len(buffer).
// Otherwise if ln >= 0, it should atleast be 128 or greater. This also
// implies that len(buffer) >= 128. Cbor object can be re-used after a
// Reset() call.
func (config *Config) NewCbor(buffer []byte, ln int) *Cbor {
	if buffer != nil && ln >= 0 && len(buffer) < 128 {
		panic("cbor buffer should atleast be 128 bytes")
	}
	if ln == -1 {
		ln = len(buffer)
	}
	return &Cbor{config: config, data: buffer, n: ln}
}

// NewJson factory to create a new Json instance. Buffer can't be nil.
// If length is less than 0, ln will be assumed as len(buffer).
// Otherwise if ln >= 0, it should atleast be 128 or greater. This also
// implies that len(buffer) >= 128. Json object can be re-used after a
// Reset() call.
func (config *Config) NewJson(buffer []byte, ln int) *Json {
	if buffer != nil && ln >= 0 && len(buffer) < 128 {
		panic("json buffer should atleast be 128 bytes")
	}
	if ln == -1 {
		ln = len(buffer)
	}
	return &Json{config: config, data: buffer, n: ln}
}

// NewCollate factor to create a new Collate instance. Buffer can't be nil.
// If length is less than 0, ln will be assumed as len(buffer).
// Otherwise if ln >= 0, it should atleast be 128 or greater. This also
// implies that len(buffer) >= 128. Collate object can be re-used after a
// Reset() call.
func (config *Config) NewCollate(buffer []byte, ln int) *Collate {
	if buffer != nil && ln >= 0 && len(buffer) < 128 {
		panic("collate buffer should atleast be 128 bytes")
	}
	if ln == -1 {
		ln = len(buffer)
	}
	return &Collate{config: config, data: buffer, n: ln}
}

// NewValue factory to create a new Value instance. Value instances are
// immutable, and can be used and re-used any number of times.
func (config *Config) NewValue(value interface{}) *Value {
	return &Value{config: config, data: value}
}

// NewJsonpointer create a instance of Jsonpointer.
func (config *Config) NewJsonpointer(path string) *Jsonpointer {
	if len(path) > config.jptrMaxlen {
		panic("jsonpointer path exceeds configured length")
	}
	jptr := &Jsonpointer{
		config:   config,
		path:     make([]byte, config.jptrMaxlen+16),
		segments: make([][]byte, config.jptrMaxseg),
	}
	for i := 0; i < config.jptrMaxseg; i++ {
		jptr.segments[i] = make([]byte, 0, 16)
	}
	n := copy(jptr.path, path)
	jptr.path = jptr.path[:n]
	return jptr
}

func (config *Config) String() string {
	return fmt.Sprintf(
		"nk:%v, ws:%v, ct:%v, arrayLenPrefix:%v, "+
			"propertyLenPrefix:%v, doMissing:%v",
		config.nk, config.ws, config.ct,
		config.arrayLenPrefix, config.propertyLenPrefix,
		config.doMissing)
}

func (nk NumberKind) String() string {
	switch nk {
	case SmartNumber:
		return "SmartNumber"
	case FloatNumber:
		return "FloatNumber"
	default:
		panic("new number-kind")
	}
}

func (ct ContainerEncoding) String() string {
	switch ct {
	case LengthPrefix:
		return "LengthPrefix"
	case Stream:
		return "Stream"
	default:
		panic("new space-kind")
	}
}
