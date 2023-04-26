package elb

type EdgeServerGroup struct {
	Backends []EdgeBackendAttribute
}

type EdgeBackendAttribute struct {
	ServerId string
	ServerIp string
	Type     string
	Port     string
	Weight   int
}

const (
	//ServerGroup
	ServerGroupDefaultType         = "ens"
	ServerGroupDefaultServerWeight = 100
	ServerGroupDefaultPort         = "0"
)
