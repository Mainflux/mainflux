package common

// errstring returns the string representation of an error.
func Errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

func GetDbDir() string {
	InitConf()
	dbDir, err := GetDataLoc()
	if err != nil {
		Log.Fatal(err)
	}
	return dbDir
}
