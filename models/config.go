package models

type ConfigModel struct {
	Discord   DiscordConfigModel   `json:"discord"`
	Storage   StorageConfigModel   `json:"storage"`
	Webserver WebserverConfigModel `json:"webserver"`
}

type DiscordConfigModel struct {
	Token           string `json:"token"`
	CooldownSeconds uint   `json:"cooldownseconds"`
	MessageLimit    uint   `json:"messageslimit"`
}

type StorageConfigModel struct {
	LifetimeSeconds uint   `json:"lifetimeseconds"`
	Location        string `json:"lcoation"`
}

type WebserverConfigModel struct {
	BindAddress   string `json:"bindaddress"`
	PublicAddress string `json:"publicaddress"`
}

var DefaultConfig = ConfigModel{
	Storage: StorageConfigModel{
		LifetimeSeconds: 30 * 60,
		Location:        "./collections",
	},
	Webserver: WebserverConfigModel{
		BindAddress:   "0.0.0.0:80",
		PublicAddress: "http://localhost:80",
	},
	Discord: DiscordConfigModel{
		CooldownSeconds: 10 * 60,
		MessageLimit:    100_000,
	},
}
