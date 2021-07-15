// +build !windows

package common

import (
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	"log/syslog"
	"os"
)

func initSyslog() {
	if "true" == os.Getenv(KuiperSyslogKey) {
		if hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, ""); err != nil {
			Log.Error("Unable to connect to local syslog daemon")
		} else {
			Log.AddHook(hook)
		}
	}
}
