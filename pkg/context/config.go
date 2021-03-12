package context

var CFG = &Config{}

type Config struct {
	Region       string `json:"region,omitempty" protobuf:"bytes,2,opt,name=region"`
	AccessKey    string `json:"accessKey,omitempty" protobuf:"bytes,2,opt,name=accessKey"`
	AccessSecret string `json:"accessSecret,omitempty" protobuf:"bytes,2,opt,name=accessSecret"`
	UID          string `json:"uid,omitempty" protobuf:"bytes,2,opt,name=uid"`
}

type Flags struct {
	LogLevel           string
	Config             string
	EnableLeaderSelect bool
}

var GlobalFlag = &Flags{}
