package common

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"github.com/benbjohnson/clock"
	"github.com/emqx/kuiper/xstream/api"
	"github.com/go-yaml/yaml"
	"github.com/keepeye/logrus-filename"
	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const (
	logFileName     = "stream.log"
	etc_dir         = "etc"
	data_dir        = "data"
	log_dir         = "log"
	plugins_dir     = "plugins"
	StreamConf      = "kuiper.yaml"
	KuiperBaseKey   = "KuiperBaseKey"
	KuiperSyslogKey = "KuiperSyslogKey"
	MetaKey         = "__meta"
)

var (
	Log             *logrus.Logger
	Config          *KuiperConf
	IsTesting       bool
	Clock           clock.Clock
	logFile         *os.File
	LoadFileType    = "relative"
	AbsoluteMapping = map[string]string{
		etc_dir:     "/etc/kuiper",
		data_dir:    "/var/lib/kuiper/data",
		log_dir:     "/var/log/kuiper",
		plugins_dir: "/var/lib/kuiper/plugins",
	}
)

func LoadConf(confName string) ([]byte, error) {
	confDir, err := GetConfLoc()
	if err != nil {
		return nil, err
	}

	file := path.Join(confDir, confName)
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return b, nil
}

type tlsConf struct {
	Certfile string `yaml:"certfile"`
	Keyfile  string `yaml:"keyfile"`
}

type KuiperConf struct {
	Basic struct {
		Debug          bool     `yaml:"debug"`
		ConsoleLog     bool     `yaml:"consoleLog"`
		FileLog        bool     `yaml:"fileLog"`
		RotateTime     int      `yaml:"rotateTime"`
		MaxAge         int      `yaml:"maxAge"`
		Ip             string   `yaml:"ip"`
		Port           int      `yaml:"port"`
		RestIp         string   `yaml:"restIp"`
		RestPort       int      `yaml:"restPort"`
		RestTls        *tlsConf `yaml:"restTls"`
		Prometheus     bool     `yaml:"prometheus"`
		PrometheusPort int      `yaml:"prometheusPort"`
		PluginHosts    string   `yaml:"pluginHosts"`
	}
	Rule api.RuleOption
	Sink struct {
		CacheThreshold    int  `yaml:"cacheThreshold"`
		CacheTriggerCount int  `yaml:"cacheTriggerCount"`
		DisableCache      bool `yaml:"disableCache""`
	}
}

func init() {
	Log = logrus.New()
	initSyslog()
	filenameHook := filename.NewHook()
	filenameHook.Field = "file"
	Log.AddHook(filenameHook)

	Log.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
		FullTimestamp:   true,
	})

	Log.Debugf("init with args %s", os.Args)
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			IsTesting = true
			break
		}
	}
	if IsTesting {
		Log.Debugf("running in testing mode")
		Clock = clock.NewMock()
	} else {
		Clock = clock.New()
	}
}

func InitConf() {
	b, err := LoadConf(StreamConf)
	if err != nil {
		Log.Fatal(err)
	}

	kc := KuiperConf{
		Rule: api.RuleOption{
			LateTol:            1000,
			Concurrency:        1,
			BufferLength:       1024,
			CheckpointInterval: 300000, //5 minutes
			SendError:          true,
		},
	}
	if err := yaml.Unmarshal(b, &kc); err != nil {
		Log.Fatal(err)
	} else {
		Config = &kc
	}
	if 0 == len(Config.Basic.Ip) {
		Config.Basic.Ip = "0.0.0.0"
	}
	if 0 == len(Config.Basic.RestIp) {
		Config.Basic.RestIp = "0.0.0.0"
	}

	if Config.Basic.Debug {
		Log.SetLevel(logrus.DebugLevel)
	}

	if Config.Basic.FileLog {
		logDir, err := GetLoc(log_dir)
		if err != nil {
			Log.Fatal(err)
		}

		file := path.Join(logDir, logFileName)
		logWriter, err := rotatelogs.New(
			file+".%Y-%m-%d_%H-%M-%S",
			rotatelogs.WithLinkName(file),
			rotatelogs.WithRotationTime(time.Hour*time.Duration(Config.Basic.RotateTime)),
			rotatelogs.WithMaxAge(time.Hour*time.Duration(Config.Basic.MaxAge)),
		)

		if err != nil {
			fmt.Println("Failed to init log file settings..." + err.Error())
			Log.Infof("Failed to log to file, using default stderr.")
		} else if Config.Basic.ConsoleLog {
			mw := io.MultiWriter(os.Stdout, logWriter)
			Log.SetOutput(mw)
		} else if !Config.Basic.ConsoleLog {
			Log.SetOutput(logWriter)
		}
	} else if Config.Basic.ConsoleLog {
		Log.SetOutput(os.Stdout)
	}
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}

func GetConfLoc() (string, error) {
	return GetLoc(etc_dir)
}

func GetDataLoc() (string, error) {
	if IsTesting {
		dataDir, err := GetLoc(data_dir)
		if err != nil {
			return "", err
		}
		d := path.Join(dataDir, "test")
		if _, err := os.Stat(d); os.IsNotExist(err) {
			err = os.MkdirAll(d, 0755)
			if err != nil {
				return "", err
			}
		}
		return d, nil
	}
	return GetLoc(data_dir)
}

func GetPluginsLoc() (string, error) {
	return GetLoc(plugins_dir)
}

func absolutePath(loc string) (dir string, err error) {
	for relDir, absoluteDir := range AbsoluteMapping {
		if strings.HasPrefix(loc, relDir) {
			dir = strings.Replace(loc, relDir, absoluteDir, 1)
			break
		}
	}
	if 0 == len(dir) {
		return "", fmt.Errorf("location %s is not allowed for absolue mode", loc)
	}
	return dir, nil
}

// GetLoc subdir must be a relative path
func GetLoc(subdir string) (string, error) {
	if "relative" == LoadFileType {
		return relativePath(subdir)
	}

	if "absolute" == LoadFileType {
		return absolutePath(subdir)
	}
	return "", fmt.Errorf("Unrecognized loading method.")
}

func relativePath(subdir string) (dir string, err error) {
	dir, err = os.Getwd()
	if err != nil {
		return "", err
	}

	if base := os.Getenv(KuiperBaseKey); base != "" {
		Log.Infof("Specified Kuiper base folder at location %s.\n", base)
		dir = base
	}
	confDir := path.Join(dir, subdir)
	if _, err := os.Stat(confDir); os.IsNotExist(err) {
		lastdir := dir
		for len(dir) > 0 {
			dir = filepath.Dir(dir)
			if lastdir == dir {
				break
			}
			confDir = path.Join(dir, subdir)
			if _, err := os.Stat(confDir); os.IsNotExist(err) {
				lastdir = dir
				continue
			} else {
				//Log.Printf("Trying to load file from %s", confDir)
				return confDir, nil
			}
		}
	} else {
		//Log.Printf("Trying to load file from %s", confDir)
		return confDir, nil
	}

	return "", fmt.Errorf("conf dir not found, please set KuiperBaseKey program environment variable correctly.")
}

func ProcessPath(p string) (string, error) {
	if abs, err := filepath.Abs(p); err != nil {
		return "", nil
	} else {
		if _, err := os.Stat(abs); os.IsNotExist(err) {
			return "", err
		}
		return abs, nil
	}
}

func ReadJsonUnmarshal(path string, ret interface{}) error {
	sliByte, err := ioutil.ReadFile(path)
	if nil != err {
		return err
	}
	err = json.Unmarshal(sliByte, ret)
	if nil != err {
		return err
	}
	return nil
}
func WriteYamlMarshal(path string, data interface{}) error {
	y, err := yaml.Marshal(data)
	if nil != err {
		return err
	}
	return ioutil.WriteFile(path, y, 0666)
}

func ReadYamlUnmarshal(path string, ret interface{}) error {
	sliByte, err := ioutil.ReadFile(path)
	if nil != err {
		return err
	}
	err = yaml.Unmarshal(sliByte, ret)
	if nil != err {
		return err
	}
	return nil
}

func UnzipTo(f *zip.File, fpath string) error {
	_, err := os.Stat(fpath)

	if f.FileInfo().IsDir() {
		// Make Folder
		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				return err
			}
		}
		return nil
	}

	if err == nil || !os.IsNotExist(err) {
		if err = os.RemoveAll(fpath); err != nil {
			return fmt.Errorf("failed to delete file %s", fpath)
		}
	}
	if _, err := os.Stat(filepath.Dir(fpath)); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}
	}

	outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		return err
	}

	rc, err := f.Open()
	if err != nil {
		return err
	}

	_, err = io.Copy(outFile, rc)

	outFile.Close()
	rc.Close()
	return err
}
