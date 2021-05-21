package framework

type configFile struct {
	YOUR_CERT_ID        string
	YOUR_LB_ID          string
	YOUR_ALC_ID         string
	YOUR_VSWITCH_ID     string
	YOUR_MASTER_ZONE_ID string
	YOUR_SLAVE_ZONE_ID  string
	YOUR_BACKEND_LABEL  string
	YOUR_RG_ID          string
}

var CONF configFile

func init() {
	CONF.YOUR_CERT_ID = "1725563258775854_16c560340da_1174296484_-802093077"
	CONF.YOUR_LB_ID = "lb-8vb67k6yu1obcx6mllzvn"
	CONF.YOUR_ALC_ID = "acl-8vbiuyiehfm2q3n8p89s4"
	CONF.YOUR_VSWITCH_ID = "vsw-8vb26y95ngam85ztcewi2"
	CONF.YOUR_MASTER_ZONE_ID = "cn-zhangjiakou-a"
	CONF.YOUR_SLAVE_ZONE_ID = "cn-zhangjiakou-b"
	CONF.YOUR_BACKEND_LABEL = "failure-domain.beta.kubernetes.io/region=cn-zhangjiakou,worker=3"
	CONF.YOUR_RG_ID = "rg-acfm2rdirl52slq"
}
