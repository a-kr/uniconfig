package uniconfig

import (
	"flag"
	"os"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

const ()

func AssertEquals(t *testing.T, actual, expected interface{}) {
	switch actual := actual.(type) {
	case string:
		if expected, ok := expected.(string); ok {
			if actual != expected {
				debug.PrintStack()
				t.Fatalf("%s != %s", actual, expected)
			}
		} else {
			t.Fatalf("Cannot compare: %v and %v (not string)", actual, expected)
		}
	case int:
		if expected, ok := expected.(int); ok {
			if actual != expected {
				debug.PrintStack()
				t.Fatalf("%d != %d", actual, expected)
			}
		} else {
			debug.PrintStack()
			t.Fatalf("Cannot compare: %v and %v (not int)", actual, expected)
		}
	case bool:
		if expected, ok := expected.(bool); ok {
			if actual != expected {
				debug.PrintStack()
				t.Fatalf("%s != %s", actual, expected)
			}
		} else {
			debug.PrintStack()
			t.Fatalf("Cannot compare: %v and %v (not bool)", actual, expected)
		}
	case time.Duration:
		if expected, ok := expected.(time.Duration); ok {
			if actual != expected {
				debug.PrintStack()
				t.Fatalf("%s != %s", actual, expected)
			}
		} else {
			t.Fatalf("Cannot compare: %v and %v (not string)", actual, expected)
		}
	default:
		debug.PrintStack()
		t.Fatalf("Cannot compare: %v and %v", actual, expected)
	}
}

type TestConfig struct {
	Debug   bool
	Count   int `help:"number of items"`
	Nested1 struct {
		A       string
		B       string
		ignored string
	}
	Nested2 struct {
		Zzz bool
	}
	Count2 int64
	TTL    time.Duration
}

func TestScanConfig(t *testing.T) {
	config := &TestConfig{}
	// some defaults
	config.Count = 42
	config.Nested1.B = "baa"
	config.TTL = time.Second

	items := ScanConfig(config)
	AssertEquals(t, len(items), 7)
	AssertEquals(t, items[0].Section, "")
	AssertEquals(t, items[0].Name, "Debug")
	AssertEquals(t, items[0].Value.Interface(), false)
	AssertEquals(t, items[0].Help, "")
	AssertEquals(t, items[1].Section, "")
	AssertEquals(t, items[1].Name, "Count")
	AssertEquals(t, items[1].Value.Interface(), 42)
	AssertEquals(t, items[1].Help, "number of items")
	AssertEquals(t, items[2].Section, "Nested1")
	AssertEquals(t, items[2].Name, "A")
	AssertEquals(t, items[2].Value.Interface(), "")
	AssertEquals(t, items[3].Section, "Nested1")
	AssertEquals(t, items[3].Name, "B")
	AssertEquals(t, items[3].Value.Interface(), "baa")
	AssertEquals(t, items[4].Section, "Nested2")
	AssertEquals(t, items[4].Name, "Zzz")
	AssertEquals(t, items[4].Value.Interface(), false)
	AssertEquals(t, items[4].Name, "Zzz")
	AssertEquals(t, items[4].Value.Interface(), false)
	AssertEquals(t, items[6].Section, "")
	AssertEquals(t, items[6].Name, "TTL")
	AssertEquals(t, items[6].Value.Interface(), time.Second)
	AssertEquals(t, items[6].Help, "")
}

func TestLoadFromEnv(t *testing.T) {
	config := TestConfig{}
	// some defaults
	config.Count = 42
	config.Nested1.B = "baa"
	items := ScanConfig(&config)

	os.Setenv("DEBUG", "true")
	os.Setenv("TTL", "10s")
	os.Setenv("NESTED1_B", "buu")
	os.Setenv("NESTED1_A", "wtf")

	// need to reset the state between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	InitFlags(items)
	LoadFromEnv(items)

	AssertEquals(t, config.Debug, true)
	AssertEquals(t, config.Count, 42)
	AssertEquals(t, config.TTL, 10*time.Second)
	AssertEquals(t, config.Nested1.A, "wtf")
	AssertEquals(t, config.Nested1.B, "buu")
	AssertEquals(t, config.Nested2.Zzz, false)
}

func TestLoadFromIni(t *testing.T) {
	config := TestConfig{}
	// some defaults
	config.Count = 42
	config.Nested1.B = "baa"
	items := ScanConfig(&config)

	// need to reset the state between tests
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	InitFlags(items)

	testIni := `
		debug = true
		count = 65535
		; this is a comment
		# also a comment
		TTL=30m

		[Nested1]
		A  = sometag

`

	dict := ParseIniFile(strings.NewReader(testIni))
	SetFromParsedIniFile(items, dict)
	AssertEquals(t, config.Debug, true)
	AssertEquals(t, config.Count, 65535)
	AssertEquals(t, config.TTL, 30*time.Minute)
	AssertEquals(t, config.Nested1.A, "sometag")
	AssertEquals(t, config.Nested1.B, "baa")
	AssertEquals(t, config.Nested2.Zzz, false)

	testIni2 := `
		debug = true
		count = 65535
		; this is a comment
		# also a comment

		[Nested1]
		A  = sometag
		unknown_parameter = must panic

`
	dict2 := ParseIniFile(strings.NewReader(testIni2))

	func() {
		defer func() {
			if err := recover(); err != nil {
			}
		}()
		SetFromParsedIniFile(items, dict2)
		t.Fatal("Test must panic with unknown parameter error")
	}()
}

func TestParseCmdline(t *testing.T) {
	args := []string{}
	configFile := GetConfigPathFromCmd(args)
	AssertEquals(t, configFile, "")
	args = []string{"--bubu", "config", "--bebe"}
	configFile = GetConfigPathFromCmd(args)
	AssertEquals(t, configFile, "")
	args = []string{"--test", "--config", "megaconfig", "--bebe"}
	configFile = GetConfigPathFromCmd(args)
	AssertEquals(t, configFile, "megaconfig")
	args = []string{"--test", "-config", "megaconfig2 42", "--bebe"}
	configFile = GetConfigPathFromCmd(args)
	AssertEquals(t, configFile, "megaconfig2 42")
	args = []string{"--test", "-config=\"4249\"", "megaconfig2 42", "--bebe"}
	configFile = GetConfigPathFromCmd(args)
	AssertEquals(t, configFile, "4249")
	args = []string{"--test", "--config=\"44991\"", "megaconfig2 42", "--bebe"}
	configFile = GetConfigPathFromCmd(args)
	AssertEquals(t, configFile, "44991")
}
