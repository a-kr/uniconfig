package uniconfig

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

var ENV_PREFIX = ""

type ConfigItem struct {
	Section string
	Name    string
	Value   reflect.Value
	Help    string
}

func (i *ConfigItem) EnvVarName() string {
	name := strings.ToUpper(i.Name)
	if i.Section == "" {
		return ENV_PREFIX + name
	}
	return ENV_PREFIX + strings.ToUpper(i.Section) + "_" + name
}

func (i *ConfigItem) CmdFlagName() string {
	name := strings.ToLower(i.Name)
	if i.Section == "" {
		return name
	}
	return strings.ToLower(i.Section) + "-" + name
}

func (i *ConfigItem) InitFlag() {
	name := i.CmdFlagName()
	switch i.Value.Kind() {
	case reflect.String:
		v := i.Value.Addr().Interface().(*string)
		flag.StringVar(v, name, *v, i.Help)
	case reflect.Int:
		v := i.Value.Addr().Interface().(*int)
		flag.IntVar(v, name, *v, i.Help)
	case reflect.Int64:
		v := i.Value.Addr().Interface().(*int64)
		flag.Int64Var(v, name, *v, i.Help)
	case reflect.Bool:
		v := i.Value.Addr().Interface().(*bool)
		flag.BoolVar(v, name, *v, i.Help)
	case reflect.Slice:
		switch i.Value.Type() {
		case intSliceType:
			v := i.Value.Addr().Interface().(*[]int)
			v1 := NewIntSlice(v)
			flag.Var(v1, name, i.Help)
		case strSliceType:
			v := i.Value.Addr().Interface().(*[]string)
			v1 := NewStrSlice(v)
			flag.Var(v1, name, i.Help)
		case floatSliceType:
			v := i.Value.Addr().Interface().(*[]float64)
			v1 := NewFloatSlice(v)
			flag.Var(v1, name, i.Help)
		default:
			log.Fatalf("Unexpected type of config entry: %v", i)
		}
	default:
		log.Fatalf("Unexpected type of config entry: %v", i)
	}
}

func ScanConfig(config interface{}) []*ConfigItem {
	items := make([]*ConfigItem, 0)

	eConfig := reflect.ValueOf(config).Elem()
	tConfig := eConfig.Type()

	for i := 0; i < tConfig.NumField(); i++ {
		f := tConfig.Field(i)

		if f.PkgPath != "" {
			continue // skip private fields
		}
		v := eConfig.Field(i)
		if f.Type.Kind() == reflect.Struct {
			for j := 0; j < f.Type.NumField(); j++ {
				ff := f.Type.Field(j)
				if ff.PkgPath != "" {
					continue // skip private fields
				}
				vv := v.Field(j)
				item := &ConfigItem{
					Section: f.Name,
					Name:    ff.Name,
					Value:   vv,
					Help:    ff.Tag.Get("help"),
				}
				items = append(items, item)
			}
			continue
		}
		item := &ConfigItem{
			Section: "",
			Name:    f.Name,
			Value:   v,
			Help:    f.Tag.Get("help"),
		}
		items = append(items, item)
	}
	return items
}

func LoadFromEnv(configItems []*ConfigItem) {
	for _, item := range configItems {
		v := os.Getenv(item.EnvVarName())
		if v != "" {
			flag.Set(item.CmdFlagName(), v)
		}
	}
}

func ParseIniFile(inifile io.Reader) map[string]string {
	scanner := bufio.NewScanner(inifile)
	result := make(map[string]string)
	// keys are stored in form of KEY or SECTION_KEY, always-uppercase
	// (so they match our convention for environment variable names)
	section := ""

	reSection := regexp.MustCompile(`^\[(.*)\]$`)
	reKeyValue := regexp.MustCompile(`^([^=]+?)\s*=\s*(.*)$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if len(line) == 0 || line[0] == ';' || line[0] == '#' {
			continue
		}
		if m := reSection.FindStringSubmatch(line); m != nil {
			section = strings.ToUpper(m[1])
			continue
		}
		if m := reKeyValue.FindStringSubmatch(line); m != nil {
			key := strings.ToUpper(m[1])
			if section != "" {
				key = section + "_" + key
			}
			value := m[2]
			result[key] = value
			continue
		}
	}
	return result
}

func GetConfigPathFromCmd(args []string) string {
	// We will search the command line for --config option.
	// We must do this manually and not via `flag` package,
	// because cmd flags must override config file params
	// (and therefore flag.Parse() must be called *after* we've read the config file)
	reArgValue := regexp.MustCompile(`^[^=]+="?(.+?)"?$`) // optionally quoted value
	for i, arg := range args {
		if (arg == "-config" || arg == "--config") && i < len(args)-1 {
			return args[i+1]
		}
		if strings.HasPrefix(arg, "-config=") || strings.HasPrefix(arg, "--config=") {
			if m := reArgValue.FindStringSubmatch(arg); m != nil {
				return m[1]
			}
		}
	}
	return ""
}

func SetFromParsedIniFile(configItems []*ConfigItem, ini map[string]string) {
	for _, item := range configItems {
		k := item.EnvVarName()
		v, ok := ini[k]
		if ok {
			flag.Set(item.CmdFlagName(), v)
			delete(ini, k)
		}
	}
	if len(ini) > 0 {
		for k, _ := range ini {
			log.Panicf("Unknown parameter in config file: %s", k)
		}
	}
}

func LoadFromConfigFile(configItems []*ConfigItem) {
	configFilename := GetConfigPathFromCmd(os.Args[1:])
	if configFilename == "" {
		return
	}
	f, err := os.Open(configFilename)
	if err != nil {
		log.Fatalf("Error loading config file %s: %s", configFilename, err)
	}
	defer f.Close()
	dict := ParseIniFile(f)
	SetFromParsedIniFile(configItems, dict)
}

func ItemsAsIniFile(configItems []*ConfigItem) string {
	sections := make(map[string][]*ConfigItem)
	for _, item := range configItems {
		sections[item.Section] = append(sections[item.Section], item)
	}
	iniLines := make([]string, 0)
	dumpSection := func(section string) {
		for _, item := range sections[section] {
			line := fmt.Sprintf("%s = %v", item.Name, item.Value.Interface())
			iniLines = append(iniLines, line)
		}
		iniLines = append(iniLines, "")
	}
	dumpSection("")
	delete(sections, "")
	for section := range sections {
		line := fmt.Sprintf("[%s]", section)
		iniLines = append(iniLines, line)
		dumpSection(section)
	}
	return strings.Join(iniLines, "\n")
}

func InitFlags(configItems []*ConfigItem) {
	flag.String("config", "", "path to configuration file") // only to provide help
	for _, item := range configItems {
		item.InitFlag()
	}
}

func LoadFromFlags(configItems []*ConfigItem) {
	flag.Parse()
}

func ConfigAsIniFile(config interface{}) string {
	items := ScanConfig(config)
	return ItemsAsIniFile(items)
}

func Load(config interface{}) {
	items := ScanConfig(config)
	InitFlags(items)
	LoadFromConfigFile(items)
	LoadFromEnv(items)
	LoadFromFlags(items)
}
