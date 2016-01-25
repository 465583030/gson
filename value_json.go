//  Copyright (c) 2015 Couchbase, Inc.

// transform golang native value into json encoded value.
// cnf: -

package gson

import "strconv"

func value2json(value interface{}, out []byte, config *Config) int {
	var err error

	if value == nil {
		return copy(out, "null")
	}

	switch v := value.(type) {
	case bool:
		if v {
			return copy(out, "true")
		}
		return copy(out, "false")

	case byte:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int8:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int16:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case uint16:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int32:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case uint32:
		out = strconv.AppendInt(out[:0], int64(v), 10)
		return len(out)

	case int:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case uint:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case int64:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case uint64:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case float32:
		out = strconv.AppendFloat(out[:0], float64(v), 'f', -1, 64)
		return len(out)

	case float64:
		out = strconv.AppendFloat(out[:0], v, 'f', -1, 64)
		return len(out)

	case string:
		out, err = encodeString(str2bytes(v), out[:0])
		if err != nil {
			panic("error encoding string")
		}
		return len(out)

	case []interface{}:
		n := 0
		out[n] = '['
		n++
		for i, x := range v {
			n += value2json(x, out[n:], config)
			if i < len(v)-1 {
				out[n] = ','
				n++
			}
		}
		out[n] = ']'
		n++
		return n

	case map[string]interface{}:
		n := 0
		out[n] = '{'
		n++

		count := len(v)
		for key := range v {
			n += value2json(key, out[n:], config)
			out[n] = ':'
			n++

			n += value2json(v[key], out[n:], config)

			count--
			if count > 0 {
				out[n] = ','
				n++
			}
		}
		out[n] = '}'
		n++
		return n

	case [][2]interface{}:
		n := 0
		out[n] = '{'
		n++

		for i, item := range v {
			n += value2json(item[0], out[n:], config)
			out[n] = ':'
			n++

			n += value2json(item[1], out[n:], config)

			if i < len(v)-1 {
				out[n] = ','
				n++
			}
		}
		out[n] = '}'
		n++
		return n
	}
	return 0
}