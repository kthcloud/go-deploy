package model

type SmVolume struct {
	Name       string `bson:"name"`
	Init       bool   `bson:"init"`
	AppPath    string `bson:"appPath"`
	ServerPath string `bson:"serverPath"`
}

type SmJob struct {
	Name    string   `bson:"name"`
	Image   string   `bson:"image"`
	Command []string `bson:"command"`
	Args    []string `bson:"args"`
}

type InitContainer struct {
	Name    string   `bson:"name"`
	Image   string   `bson:"image"`
	Command []string `bson:"command"`
	Args    []string `bson:"args"`
}
