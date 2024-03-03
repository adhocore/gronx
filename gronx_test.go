package gronx

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"
)

type Case struct {
	Expr   string `json:"expr"`
	Ref    string `json:"ref"`
	Expect bool   `json:"expect"`
	Next   string `json:"next"`
}

func (test Case) run(gron Gronx) (bool, error) {
	if test.Ref == "" {
		return gron.IsDue(test.Expr)
	}

	ref, err := time.Parse(FullDateFormat, test.Ref)
	abort(err)

	return gron.IsDue(test.Expr, ref)
}

func TestNormalize(t *testing.T) {
	tests := map[string]string{
		"*   *  *\t*\n*":        "* * * * *",
		"* * * * * 2021":        "* * * * * 2021",
		"@hourly":               "0 * * * *",
		"0 0 JAN,feb * sun,MON": "0 0 1,2 * 0,1",
	}

	for expr, expect := range tests {
		t.Run("normalize "+expr, func(t *testing.T) {
			actual := strings.Join(normalize(expr), " ")

			if expect != actual {
				t.Errorf("expected %v, got %v", expect, actual)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	gron := New()

	t.Run("is valid", func(t *testing.T) {
		if !gron.IsValid("5,10-20/4,55 * * * *") {
			t.Errorf("expected true, got false")
		}
		if !gron.IsValid("00 * * * *") {
			t.Errorf("expected true, got false")
		}
		if !gron.IsValid("* 00 * * *") {
			t.Errorf("expected true, got false")
		}
	})

	t.Run("is not valid", func(t *testing.T) {
		if gron.IsValid("A-B * * * *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("60 * * * *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("* 30 * * *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("* * 99 * *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("* * * 13 *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("* * * * 8") {
			t.Errorf("expected false, got true")
		}

		if gron.IsValid("60-65 * * * *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("* 24-28/2 * * *") {
			t.Errorf("expected false, got true")
		}
		if gron.IsValid("* * * *") {
			t.Errorf("expected false, got true")
		}
	})

}

func TestAddTag(t *testing.T) {
	t.Run("add good tag", func(t *testing.T) {
		err := AddTag("@2s", "*/2 * * * * *")
		if err != nil {
			t.Error("expected nil, got err")
		}

		expr, ok := expressions["@2s"]
		if !ok {
			t.Error("expected true, got false")
		}

		if expr != "*/2 * * * * *" {
			t.Error("expected */2 * * * * *")
		}
	})

	t.Run("add conflict tag", func(t *testing.T) {
		err := AddTag("@2s", "*/2 * * * * *")
		if err == nil {
			t.Error("expected err, got nil")
		}
	})

	t.Run("add wrong tag", func(t *testing.T) {
		err := AddTag("@3s", "* * * *")
		if err == nil {
			t.Error("expected err, got nil")
		}
	})
}

func TestIsDue(t *testing.T) {
	gron := New()

	t.Run("seconds precision", func(t *testing.T) {
		expr := "*/2 * * * * *"
		ref, _ := time.Parse(FullDateFormat, "2020-02-02 02:02:04")
		due, _ := gron.IsDue(expr, ref)
		if !due {
			t.Errorf("%s should be due on %s", expr, ref)
		}

		due, _ = gron.IsDue(expr, ref.Add(time.Second))
		if due {
			t.Errorf("%s should be due on %s", expr, ref)
		}
	})

	for i, test := range testcases() {
		t.Run(fmt.Sprintf("is due #%d=%s", i, test.Expr), func(t *testing.T) {
			actual, _ := test.run(gron)

			if actual != test.Expect {
				t.Errorf("expected %v, got %v", test.Expect, actual)
			}
		})
	}

	for i, test := range errcases() {
		t.Run(fmt.Sprintf("is due err #%d=%s", i, test.Expr), func(t *testing.T) {
			actual, err := test.run(gron)

			if actual != test.Expect {
				t.Errorf("expected %v, got %v", test.Expect, actual)
			}
			if err == nil {
				t.Errorf("expected error, got nil")
			}
		})
	}
}

func TestValueByPos(t *testing.T) {
	t.Run("valueByPos 7", func(t *testing.T) {
		if actual := valueByPos(time.Now(), 7); actual != 0 {
			t.Errorf("expected 0, got %v", actual)
		}
	})
}

func testcases() []Case {
	return []Case{
		{"@always", "2021-04-19 12:54:00", true, "2021-04-19 12:55:00"},
		{"* * * * * 2018", "2022-01-02 15:04:00", false, "err"},
		{"* * * * * 2018", "2021-04-19 12:54:00", false, "err"},
		{"@5minutes", "2017-05-10 02:30:00", true, "2017-05-10 02:35:00"},
		{"* * 7W * *", "2017-10-15 20:00:00", false, "2017-11-07 00:00:00"},
		{"*/2 */2 * * *", "2015-08-10 21:47:00", false, "2015-08-10 22:00:00"},
		{"* * * * *", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"* * * * * ", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"*  *  *  *  *", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"*	*	*	*	*", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"*		* *	* *", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"* 20,21,22 * * *", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"* 20,22 * * *", "2015-08-10 21:50:00", false, "2015-08-10 22:00:00"},
		{"* 5,21-22 * * *", "2015-08-10 21:50:00", true, "2015-08-10 21:51:00"},
		{"7-9 * */9 * *", "2015-08-10 22:02:00", false, "2015-08-10 22:07:00"},
		{"7-9 * */9 * *", "2015-08-11 22:02:00", false, "2015-08-19 00:07:00"},
		{"1 * * * 7", "2015-08-10 21:47:00", false, "2015-08-16 00:01:00"},
		{"47 21 * * *", "2015-08-10 21:47:00", true, "2015-08-11 21:47:00"},
		{"00 * * * *", "2023-07-21 12:30:00", false, "2023-07-21 13:00:00"},
		{"0 00 * * *", "2023-07-21 12:30:00", false, "2023-07-22 00:00:00"},
		{"0 000 * * *", "2023-07-21 12:30:00", false, "2023-07-22 00:00:00"},
		{"* * * * 0", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"* * * * 7", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"* * * * 1", "2011-06-15 23:09:00", false, "2011-06-20 00:00:00"},
		{"0 0 * * MON,SUN", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"0 0 * * 1,7", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"0 0 * * 0-4", "2011-06-15 23:09:00", false, "2011-06-16 00:00:00"},
		{"0 0 * * 7-4", "2011-06-15 23:09:00", false, "2011-06-16 00:00:00"},
		{"0 0 * * 4-7", "2011-06-15 23:09:00", false, "2011-06-16 00:00:00"},
		{"0 0 * * 7-3", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"0 0 * * 3-7", "2011-06-15 23:09:00", false, "2011-06-16 00:00:00"},
		{"0 0 * * 3-7", "2011-06-18 23:09:00", false, "2011-06-22 00:00:00"},
		{"0 0 * * 2-7", "2011-06-20 23:09:00", false, "2011-06-21 00:00:00"},
		{"0 0 * * 0,2-6", "2011-06-20 23:09:00", false, "2011-06-21 00:00:00"},
		{"0 0 * * 2-7", "2011-06-18 23:09:00", false, "2011-06-21 00:00:00"},
		{"0 0 * * 4-7", "2011-07-19 00:00:00", false, "2011-07-21 00:00:00"},
		{"0-12/4 * * * *", "2011-06-20 12:04:00", true, "2011-06-20 12:08:00"},
		{"0-10/2 * * * *", "2011-06-20 12:12:00", false, "2011-06-20 13:00:00"},
		{"4-59/2 * * * *", "2011-06-20 12:04:00", true, "2011-06-20 12:06:00"},
		{"4-59/2 * * * *", "2011-06-20 12:06:00", true, "2011-06-20 12:08:00"},
		{"4-59/3 * * * *", "2011-06-20 12:06:00", false, "2011-06-20 12:07:00"},
		{"0 0 * * 0,2-6", "2011-06-20 23:09:00", false, "2011-06-21 00:00:00"},
		{"0 0 1 1 0", "2011-06-15 23:09:00", false, "2012-01-01 00:00:00"},
		{"0 0 1 JAN 0", "2011-06-15 23:09:00", false, "2012-01-01 00:00:00"},
		{"0 0 1 * 0", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"0 0 L * *", "2011-07-15 00:00:00", false, "2011-07-31 00:00:00"},
		{"0 0 2W * *", "2011-07-01 00:00:00", true, "2011-08-02 00:00:00"},
		{"0 0 1W * *", "2011-05-01 00:00:00", false, "2011-05-02 00:00:00"},
		{"0 0 1W * *", "2011-07-01 00:00:00", true, "2011-08-01 00:00:00"},
		{"0 0 3W * *", "2011-07-01 00:00:00", false, "2011-07-04 00:00:00"},
		{"0 0 16W * *", "2011-07-01 00:00:00", false, "2011-07-15 00:00:00"},
		{"0 0 28W * *", "2011-07-01 00:00:00", false, "2011-07-28 00:00:00"},
		{"0 0 30W * *", "2011-07-01 00:00:00", false, "2011-07-29 00:00:00"},
		// {"0 0 31W * *", "2011-07-01 00:00:00", false, "2011-07-29 00:00:00"},
		{"* * * * * 2012", "2011-05-01 00:00:00", false, "2012-01-01 00:00:00"},
		{"* * * * 5L", "2011-07-01 00:00:00", false, "2011-07-29 00:00:00"},
		{"* * * * 6L", "2011-07-01 00:00:00", false, "2011-07-30 00:00:00"},
		{"* * * * 7L", "2011-07-01 00:00:00", false, "2011-07-31 00:00:00"},
		{"* * * * 1L", "2011-07-24 00:00:00", false, "2011-07-25 00:00:00"},
		{"* * * * TUEL", "2011-07-24 00:00:00", false, "2011-07-26 00:00:00"},
		{"* * * 1 5L", "2011-12-25 00:00:00", false, "2012-01-27 00:00:00"},
		{"* * * * 5#2", "2011-07-01 00:00:00", false, "2011-07-08 00:00:00"},
		{"* * * * 5#1", "2011-07-01 00:00:00", true, "2011-07-01 00:01:00"},
		{"* * * * 3#4", "2011-07-01 00:00:00", false, "2011-07-27 00:00:00"},
		{"0 0 * * 1#1", "2009-10-23 00:00:00", false, "2009-11-02 00:00:00"},
		{"0 0 * * 1#1", "2009-11-23 00:00:00", false, "2009-12-07 00:00:00"},
		{"5/0 * * * *", "2021-04-19 12:54:00", false, "2018-08-13 00:25:00"},
		{"5/20 * * * *", "2018-08-13 00:24:00", false, "2018-08-13 00:25:00"},
		{"5/20 * * * *", "2018-08-13 00:45:00", true, "2018-08-13 01:05:00"},
		{"5-11/4 * * * *", "2018-08-13 00:03:00", false, "2018-08-13 00:05:00"},
		{"0 0 L * 0", "2011-06-15 23:09:00", false, "2011-06-19 00:00:00"},
		{"3-59/15 6-12 */15 1 2-5", "2017-01-08 00:00:00", false, "2017-01-31 06:03:00"},
		{"* * * * MON-FRI", "2017-01-08 00:00:00", false, "2017-01-09 00:00:00"},
		{"* * * * TUE", "2017-01-08 00:00:00", false, "2017-01-10 00:00:00"},
		{"0 1 15 JUL mon,Wed,FRi", "2019-11-14 00:00:00", false, "2020-07-01 01:00:00"},
		{"0 1 15 jul mon,Wed,FRi", "2019-11-14 00:00:00", false, "2020-07-01 01:00:00"},
		{"@weekly", "2019-11-14 00:00:00", false, "2019-11-17 00:00:00"},
		{"@weekly", "2019-11-14 00:00:00", false, "2019-11-17 00:00:00"},
		{"@weekly", "2019-11-14 00:00:00", false, "2019-11-17 00:00:00"},
		{"0 12 * * ?", "2020-08-20 00:00:00", false, "2020-08-20 12:00:00"},
		{"0 12 ? * *", "2020-08-20 00:00:00", false, "2020-08-20 12:00:00"},
		{"* ? * ? * *", "2020-08-20 00:00:00", true, "2020-08-20 00:00:01"},
		{"* * ? * * * */2", "2021-08-20 00:00:00", false, "2022-01-01 00:00:00"},
		{"* * * * * * *", "2021-08-20 00:00:00", true, "2021-08-20 00:00:01"},
		{"* * * * * * 2023-2099", "2021-08-20 00:00:00", false, "2023-01-01 00:00:00"},
		{"30 9 L */3 *", "2023-04-23 09:30:00", false, "2023-04-30 09:30:00"},
		{"30 9 L */3 *", "2023-05-01 09:30:00", false, "2023-07-31 09:30:00"},
	}
}

func errcases() []Case {
	return []Case{
		{"* * * *", "", false, ""},
		{"* * * * * * * *", "", false, ""},
		{"- * * * *", "2011-07-01 00:01:00", false, ""},
		{"/ * * * *", "2011-07-01 00:01:00", false, ""},
		{"Z/Z * * * *", "2011-07-01 00:01:00", false, ""},
		{"Z/0 * * * *", "2011-07-01 00:01:00", false, ""},
		{"Z-10 * * * *", "2011-07-01 00:01:00", false, ""},
		{"1-Z * * * *", "2011-07-01 00:01:00", false, ""},
		{"1-Z/2 * * * *", "2011-07-01 00:01:00", false, ""},
		{"Z-Z/2 * * * *", "2011-07-01 00:01:00", false, ""},
		{"* * 0 * *", "2011-07-01 00:01:00", false, ""},
		{"* * * W * *", "", false, ""},
		{"* * * ZW * *", "", false, ""},
		{"* * * * 4W", "2011-07-01 00:00:00", false, ""},
		{"* * * 1L *", "2011-07-01 00:00:00", false, ""},
		{"* * * * * ZL", "", false, ""},
		{"* * * * * Z#", "", false, ""},
		{"* * * * * 1#Z", "", false, ""},
	}
}

func abort(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
