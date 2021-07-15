package common

type RPCArgDesc struct {
	Name, Json string
}

type PluginDesc struct {
	RPCArgDesc
	Type int
	Stop bool
}
