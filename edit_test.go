package main

import (
	"reflect"
	"testing"
)

func TestEdit(t *testing.T) {
	testtab := []struct {
		dot      Range
		filename string
		expr     string
		expected string
	}{

		// 0
		{Range{0, 0}, "test", "a/junk", "junkThis is a\nshort text\nto try addressing\n"},
		{Range{7, 12}, "test", "a/junk", "This is a\nshjunkort text\nto try addressing\n"},
		{Range{0, 0}, "test", "/This/a/junk", "Thisjunk is a\nshort text\nto try addressing\n"},
		{Range{0, 0}, "test", "/^/a/junk", "This is a\njunkshort text\nto try addressing\n"},
		{Range{0, 0}, "test", "/$/a/junk", "This is ajunk\nshort text\nto try addressing\n"},

		// 4
		{Range{0, 0}, "test", "i/junk", "junkThis is a\nshort text\nto try addressing\n"},
		{Range{2, 6}, "test", "i/junk", "Thjunkis is a\nshort text\nto try addressing\n"},
		{Range{0, 0}, "test", "/text/i/junk", "This is a\nshort junktext\nto try addressing\n"},

		// Don't know how to automate testing of 'b'

		// c
		// 7
		{Range{0, 0}, "test", "c/junk", "junkThis is a\nshort text\nto try addressing\n"},
		{Range{2, 6}, "test", "c/junk", "Thjunks a\nshort text\nto try addressing\n"},
		{Range{0, 0}, "test", "/text/c/junk", "This is a\nshort junk\nto try addressing\n"},

		// d
		// 10
		{Range{0, 0}, "test", "d", "This is a\nshort text\nto try addressing\n"},
		{Range{2, 6}, "test", "d", "Ths a\nshort text\nto try addressing\n"},
		{Range{0, 0}, "test", "/text/d", "This is a\nshort \nto try addressing\n"},

		// e - Don't know how to test e

		// f - Don't know how to test f

		// g/v
		{Range{0, 0}, "test", "g/This/d", "This is a\nshort text\nto try addressing\n"},
		{Range{0, 12}, "test", "g/This/d", "ort text\nto try addressing\n"},
		{Range{0, 3}, "test", "v/This/d", "s is a\nshort text\nto try addressing\n"},
		{Range{0, 12}, "test", "v/This/d", "This is a\nshort text\nto try addressing\n"},

		// m/t
		// 17
		{Range{0, 4}, "test", "m/try", " is a\nshort text\nto tryThis addressing\n"},
		{Range{0, 3}, "test", "t/try", "This is a\nshort text\nto tryThi addressing\n"},
	}

	buf := make([]rune, 8192)

	for i, test := range testtab {
		w := NewWindow().initHeadless(nil)
		w.body.Insert(0, []rune("This is a\nshort text\nto try addressing\n"), true)
		w.body.SetQ0(test.dot.q0)
		w.body.SetQ1(test.dot.q1)
		editcmd(&w.body, []rune(test.expr))
		// Normally the edit log is applied in allupdate, but we don't have
		// all the window machinery, so we apply it by hand.
		w.body.file.elog.Apply(&w.body)
		n, _ := w.body.ReadB(0, buf[:])
		if string(buf[:n]) != test.expected {
			t.Errorf("test %d: TestAppend expected \n%v\nbut got \n%v\n", i, test.expected, string(buf[:n]))
		}
	}
}

func TestCollecttoken(t *testing.T) {
	tt := []struct {
		cmd []rune
		end string
		out string
	}{
		{[]rune(" foo bar\t\n"), linex, " foo bar\t"},
		{[]rune(" foo bar\t\nquux"), linex, " foo bar\t"},
		{[]rune(" αβγ テスト\t\n世界"), linex, " αβγ テスト\t"},
		{[]rune(" foo bar\t\n"), wordx, " foo bar"},
		{[]rune(" foo bar\t\nquux"), wordx, " foo bar"},
		{[]rune(" αβγ テスト\t\n世界"), wordx, " αβγ テスト"},
	}
	for _, tc := range tt {
		cmdstartp = tc.cmd
		cmdp = 0
		out := collecttoken(tc.end)
		if out != tc.out {
			t.Errorf("collecttoken(%q) of command %q is %q; exptected %q",
				tc.end, tc.cmd, out, tc.out)
		}
	}
}

func TestSimpleaddr(t *testing.T) {
	tt := []struct {
		ok   bool
		cmd  []rune
		addr *Addr
	}{
		{true, nil, nil},
		{true, []rune{}, nil},
		{true, []rune("\n"), nil},
		{true, []rune("#123\n"), &Addr{typ: '#', num: 123}},
		{true, []rune("#\n"), &Addr{typ: '#', num: 1}},
		{true, []rune("42\n"), &Addr{typ: 'l', num: 42}},
		{true, []rune("1234567890\n"), &Addr{typ: 'l', num: 1234567890}},
		{true, []rune("/abc\n"), &Addr{typ: '/', re: "abc"}},
		{true, []rune("/abc/\n"), &Addr{typ: '/', re: "abc"}},
		{true, []rune(`/a\/bc/` + "\n"), &Addr{typ: '/', re: "a/bc"}},
		{true, []rune(`/a\nbc/` + "\n"), &Addr{typ: '/', re: `a\nbc`}},
		{true, []rune(`/a\\bc/` + "\n"), &Addr{typ: '/', re: `a\\bc`}},
		{true, []rune("?abc\n"), &Addr{typ: '?', re: "abc"}},
		{true, []rune("?abc?\n"), &Addr{typ: '?', re: "abc"}},
		{true, []rune(`?a\?bc?` + "\n"), &Addr{typ: '?', re: "a?bc"}},
		{true, []rune(`?a\nbc?` + "\n"), &Addr{typ: '?', re: `a\nbc`}},
		{true, []rune(`?a\\bc?` + "\n"), &Addr{typ: '?', re: `a\\bc`}},
		{true, []rune(`"abc` + "\n"), &Addr{typ: '"', re: "abc"}},
		{true, []rune(`"abc"` + "\n"), &Addr{typ: '"', re: "abc"}},
		{true, []rune(".\n"), &Addr{typ: '.'}},
		{true, []rune("$\n"), &Addr{typ: '$'}},
		{true, []rune("+\n"), &Addr{typ: '+'}},
		{true, []rune("-\n"), &Addr{typ: '-'}},
		{true, []rune("'\n"), &Addr{typ: '\''}},
		{true, []rune("abc\n"), nil},
		{false, []rune("42.\n"), nil},
		{false, []rune("42$\n"), nil},
		{false, []rune("42'\n"), nil},
		{false, []rune("42\"\n"), nil},
		{false, []rune(`"abc" "cdf" "efg"` + "\n"), nil},
		{true, []rune("\"abc\" 42\n"), &Addr{
			typ: '"', re: "abc", next: &Addr{
				typ: 'l', num: 42,
			}}},
		{true, []rune(".42\n"), &Addr{
			typ: '.', next: &Addr{
				typ: '+', next: &Addr{
					typ: 'l', num: 42,
				}}}},
		{true, []rune("42/abc/\n"), &Addr{
			typ: 'l', num: 42, next: &Addr{
				typ: '+', next: &Addr{
					typ: '/', re: "abc",
				}}}},
		{true, []rune("42/abc/\n"), &Addr{
			typ: 'l', num: 42, next: &Addr{
				typ: '+', next: &Addr{
					typ: '/', re: "abc",
				}}}},
		{true, []rune("+/abc/\n"), &Addr{typ: '+', next: &Addr{typ: '/', re: "abc"}}},
		{true, []rune("-/abc/\n"), &Addr{typ: '-', next: &Addr{typ: '/', re: "abc"}}},
		{true, []rune(".+\n"), &Addr{typ: '.', next: &Addr{typ: '+', num: 0}}},
		{true, []rune(".-\n"), &Addr{typ: '.', next: &Addr{typ: '-', num: 0}}},
	}
	for _, tc := range tt {
		cmdstartp = tc.cmd
		cmdp = 0
		addr, err := simpleaddr()
		if tc.ok && err != nil {
			t.Errorf("simple address %q returned error %v", tc.cmd, err)
			continue
		}
		if !tc.ok && err == nil {
			t.Errorf("simple address %q returned nil error", tc.cmd)
			continue
		}
		if !reflect.DeepEqual(addr, tc.addr) {
			t.Errorf("bad parse result for address %q:\n"+
				"     got: %v\n"+
				"expected: %v",
				tc.cmd, addr, tc.addr)
		}
	}
}
