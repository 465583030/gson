// transform collated value into json encoding.
// cnf: SpaceKind, NumberKind, doMissing, arrayLenPrefix, propertyLenPrefix

package gson

import "strconv"

func json2collate(txt string, code []byte, config *Config) (string, int) {
	txt = skipWS(txt, config.ws)
	if len(txt) < 1 {
		panic("collate scanner jsonEmpty")
	}

	n := 0

	if digitCheck[txt[0]] == 1 {
		code[n] = TypeNumber
		n++
		m, remtxt := jsonnum2collate(txt, code[n:], config.nk)
		n += m
		code[n] = Terminator
		n++
		return remtxt, n
	}

	switch txt[0] {
	case 'n':
		if len(txt) >= 4 && txt[:4] == "null" {
			code[n], code[n+1] = TypeNull, Terminator
			return txt[4:], n + 2
		}
		panic("collate scanner expectedNil")

	case 't':
		if len(txt) >= 4 && txt[:4] == "true" {
			code[n], code[n+1] = TypeTrue, Terminator
			return txt[4:], n + 2
		}
		panic("collate scanner expectedTrue")

	case 'f':
		if len(txt) >= 5 && txt[:5] == "false" {
			code[n], code[n+1] = TypeFalse, Terminator
			return txt[5:], n + 2
		}
		panic("collate scanner expectedFalse")

	case '"':
		scratchi := config.pools.stringPool.Get()
		scratch := scratchi.([]byte)
		defer config.pools.stringPool.Put(scratchi)

		txt, p := scanString(txt, scratch)
		if config.doMissing && MissingLiteral.Equal(bytes2str(scratch[:p])) {
			code[n], code[n+1] = TypeMissing, Terminator
			return txt, n + 2
		}
		code[n] = TypeString
		n++
		n += suffixEncodeString(scratch[:p], code[n:])
		code[n] = Terminator
		n++
		return txt, n

	case '[':
		var x int

		code[n] = TypeArray
		n++
		nn, nnn, ln := n, n, 0
		if config.arrayLenPrefix {
			nn, nnn = (nn + 32), (nnn + 32) // prealloc space for Len encoding
		}

		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("collate scanner expectedCloseArray")

		} else if txt[0] != ']' {
			for {
				txt, x = json2collate(txt, code[nnn:], config)
				nnn += x
				ln++
				if txt = skipWS(txt, config.ws); len(txt) == 0 {
					panic("gson scanner expectedCloseArray")
				} else if txt[0] == ',' {
					txt = skipWS(txt[1:], config.ws)
				} else if txt[0] == ']' {
					break
				} else {
					panic("collate scanner expectedCloseArray")
				}
			}
		}
		if config.arrayLenPrefix {
			n += collateLength(ln, code[n:])
			n += copy(code[n:], code[nn:nnn])
		} else {
			n = nnn
		}
		code[n] = Terminator
		n++
		return txt[1:], n

	case '{':
		var x, p, ln int

		code[n] = TypeObj
		n++

		altcodei := config.pools.codepool.Get()
		altcode := altcodei.([]byte)
		defer config.pools.codepool.Put(altcodei)

		refsi := config.pools.keypool.Get()
		refs := refsi.(kvrefs)
		defer config.pools.keypool.Put(refsi)

		if txt = skipWS(txt[1:], config.ws); len(txt) == 0 {
			panic("collate scanner expectedCloseobject")
		} else if txt[0] != '}' && txt[0] != '"' {
			panic("collate scanner expectedKey")
		} else if txt[0] != '}' {
			for {
				// NOTE: empty string is also a valid key
				txt, x = scanString(txt, altcode[p:])
				if txt = skipWS(txt, config.ws); len(txt) == 0 || txt[0] != ':' {
					panic("collate scanner expectedColon")
				}
				key := bytes2str(altcode[p : p+x])
				p += x

				txt = skipWS(txt[1:], config.ws)
				txt, x = json2collate(txt, altcode[p:], config)
				refs[ln] = kvref{key, altcode[p : p+x]}
				p += x
				ln++

				if txt = skipWS(txt, config.ws); len(txt) == 0 {
					panic("collate scanner expectedCloseobject")
				} else if txt[0] == ',' {
					txt = skipWS(txt[1:], config.ws)
				} else if txt[0] == '}' {
					break
				} else {
					panic("collate scanner expectedCloseobject")
				}
			}

			(refs[:ln]).sort()
		}
		if config.propertyLenPrefix {
			n += collateLength(ln, code[n:])
		}
		for j := 0; j < ln; j++ {
			kv := refs[j]
			n += collateString(kv.key, code[n:], config) // encode key
			n += copy(code[n:], kv.code)
		}
		code[n] = Terminator
		n++
		return txt[1:], n
	}
	panic("collate scanner expectedToken")
}

func jsonnum2collate(txt string, code []byte, nk NumberKind) (int, string) {
	s, e, l := 0, 1, len(txt)
	if len(txt) > 1 {
		for ; e < l && intCheck[txt[e]] == 1; e++ {
		}
	}
	switch nk {
	case FloatNumber:
		f, err := strconv.ParseFloat(txt[s:e], 64)
		if err != nil {
			panic(err)
		}
		n := collateFloat64(f, code)
		return n, txt[e:]

	case SmartNumber:
		if i, err := strconv.ParseInt(txt[s:e], 10, 64); err == nil {
			n := collateInt64(i, code)
			return n, txt[e:]
		} else if ui, err := strconv.ParseUint(txt[s:e], 10, 64); err == nil {
			n := collateUint64(ui, code)
			return n, txt[e:]
		}
		f, err := strconv.ParseFloat(txt[s:e], 64)
		if err != nil {
			panic(err)
		}
		n := collateFloat64(f, code)
		return n, txt[e:]
	}
	panic("unreachable code")
}
