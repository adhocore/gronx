package gronx

import (
	"log"
	"strings"
	"testing"
	"time"
)

type Case struct {
	Expr   string `json:"expr"`
	Ref    string `json:"ref"`
	Expect bool   `json:"expect"`
}

func (test Case) run(gron Gronx) (bool, error) {
	if test.Ref == "" {
		return gron.IsDue(test.Expr)
	}

	ref, err := time.Parse("2006-01-02 15:04:05", test.Ref)
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
		if !gron.IsValid("5,55 * * * *") {
			t.Errorf("expected false, got true")
		}
	})

	t.Run("is not valid", func(t *testing.T) {
		if gron.IsValid("A-B * * * *") {
			t.Errorf("expected true, got false")
		}
	})
}

func TestIsDue(t *testing.T) {
	gron := New()

	for _, test := range testcases() {
		t.Run("is due "+test.Expr, func(t *testing.T) {
			actual, _ := test.run(gron)

			if actual != test.Expect {
				t.Errorf("expected %v, got %v", test.Expect, actual)
			}
		})
	}

	for _, test := range errcases() {
		t.Run("is due "+test.Expr, func(t *testing.T) {
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
		{"@always", "2021-04-19 12:54:09", true},
		{"* * * * * 2018", "", false},
		{"* * * * * 2018", "2021-04-19 12:54:09", false},
		{"@5minutes", "2017-05-10 02:30:00", true},
		{"* * 7W * *", "2017-10-15 20:00:00", false},
		{"*/2 */2 * * *", "2015-08-10 21:47:27", false},
		{"* * * * *", "2015-08-10 21:50:37", true},
		{"* * * * * ", "2015-08-10 21:50:37", true},
		{"*  *  *  *  *", "2015-08-10 21:50:37", true},
		{"*	*	*	*	*", "2015-08-10 21:50:37", true},
		{"*		* *	* *", "2015-08-10 21:50:37", true},
		{"* 20,21,22 * * *", "2015-08-10 21:50:00", true},
		{"* 20,22 * * *", "2015-08-10 21:50:00", false},
		{"* 5,21-22 * * *", "2015-08-10 21:50:00", true},
		{"7-9 * */9 * *", "2015-08-10 22:02:33", false},
		{"1 * * * 7", "2015-08-10 21:47:27", false},
		{"47 21 * * *", "2015-08-10 21:47:30", true},
		{"* * * * 0", "2011-06-15 23:09:00", false},
		{"* * * * 7", "2011-06-15 23:09:00", false},
		{"* * * * 1", "2011-06-15 23:09:00", false},
		{"0 0 * * MON,SUN", "2011-06-15 23:09:00", false},
		{"0 0 * * 1,7", "2011-06-15 23:09:00", false},
		{"0 0 * * 0-4", "2011-06-15 23:09:00", false},
		{"0 0 * * 7-4", "2011-06-15 23:09:00", false},
		{"0 0 * * 4-7", "2011-06-15 23:09:00", false},
		{"0 0 * * 7-3", "2011-06-15 23:09:00", false},
		{"0 0 * * 3-7", "2011-06-15 23:09:00", false},
		{"0 0 * * 3-7", "2011-06-18 23:09:00", false},
		{"0 0 * * 2-7", "2011-06-20 23:09:00", false},
		{"0 0 * * 0,2-6", "2011-06-20 23:09:00", false},
		{"0 0 * * 2-7", "2011-06-18 23:09:00", false},
		{"0 0 * * 4-7", "2011-07-19 00:00:00", false},
		{"0-12/4 * * * *", "2011-06-20 12:04:00", true},
		{"4-59/2 * * * *", "2011-06-20 12:04:00", true},
		{"4-59/2 * * * *", "2011-06-20 12:06:00", true},
		{"4-59/3 * * * *", "2011-06-20 12:06:00", false},
		{"0 0 * * 0,2-6", "2011-06-20 23:09:00", false},
		{"0 0 1 1 0", "2011-06-15 23:09:00", false},
		{"0 0 1 JAN 0", "2011-06-15 23:09:00", false},
		{"0 0 1 * 0", "2011-06-15 23:09:00", false},
		{"0 0 L * *", "2011-07-15 00:00:00", false},
		{"0 0 2W * *", "2011-07-01 00:00:00", true},
		{"0 0 1W * *", "2011-05-01 00:00:00", false},
		{"0 0 1W * *", "2011-07-01 00:00:00", true},
		{"0 0 3W * *", "2011-07-01 00:00:00", false},
		{"0 0 16W * *", "2011-07-01 00:00:00", false},
		{"0 0 28W * *", "2011-07-01 00:00:00", false},
		{"0 0 30W * *", "2011-07-01 00:00:00", false},
		{"0 0 31W * *", "2011-07-01 00:00:00", false},
		{"* * * * * 2012", "2011-05-01 00:00:00", false},
		{"* * * * 5L", "2011-07-01 00:00:00", false},
		{"* * * * 6L", "2011-07-01 00:00:00", false},
		{"* * * * 7L", "2011-07-01 00:00:00", false},
		{"* * * * 1L", "2011-07-24 00:00:00", false},
		{"* * * * TUEL", "2011-07-24 00:00:00", false},
		{"* * * 1 5L", "2011-12-25 00:00:00", false},
		{"* * * * 5#2", "2011-07-01 00:00:00", false},
		{"* * * * 5#1", "2011-07-01 00:00:00", true},
		{"* * * * 3#4", "2011-07-01 00:00:00", false},
		{"5/0 * * * *", "2021-04-19 12:54:09", false},
		{"5/20 * * * *", "2018-08-13 00:24:00", false},
		{"5/20 * * * *", "2018-08-13 00:45:00", true},
		{"5-11/4 * * * *", "2018-08-13 00:03:00", false},
		{"0 0 L * 0", "2011-06-15 23:09:00", false},
		{"3-59/15 6-12 */15 1 2-5", "2017-01-08 00:00:00", false},
		{"* * * * MON-FRI", "2017-01-08 00:00:00", false},
		{"* * * * TUE", "2017-01-08 00:00:00", false},
		{"0 1 15 JUL mon,Wed,FRi", "2019-11-14 00:00:00", false},
		{"0 1 15 jul mon,Wed,FRi", "2019-11-14 00:00:00", false},
		{"@weekly", "2019-11-14 00:00:00", false},
		{"@weekly", "2019-11-14 00:00:00", false},
		{"@weekly", "2019-11-14 00:00:00", false},
		{"0 12 * * ?", "2020-08-20 00:00:00", false},
		{"0 12 ? * *", "2020-08-20 00:00:00", false},
	}
}

func errcases() []Case {
	return []Case{
		{"* * * *", "", false},
		{"* * * * * * *", "", false},
		{"- * * * *", "2011-07-01 00:01:00", false},
		{"/ * * * *", "2011-07-01 00:01:00", false},
		{"Z/Z * * * *", "2011-07-01 00:01:00", false},
		{"Z-Z/2 * * * *", "2011-07-01 00:01:00", false},
		{"* * W * *", "", false},
		{"* * ZW * *", "", false},
		{"* * * * 4W", "2011-07-01 00:00:00", false},
		{"* * * 1L *", "2011-07-01 00:00:00", false},
		{"* * * * ZL", "", false},
		{"* * * * Z#", "", false},
		{"* * * * 1#Z", "", false},
	}
}

func abort(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
