package domain

import (
	"encoding/json"
	"os"
	"os/user"
	"regexp"
	"runtime"
	"time"

	"github.com/denisbrodbeck/machineid"
	jsoniter "github.com/json-iterator/go"
	"github.com/mantil-io/mantil/kit/gz"
)

/*
 gomodifytags -all -add-tags short -add-options short=omitempty -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags bson  -add-options bson=omitempty  -transform=camelcase -w -file event.go
 gomodifytags -all -add-tags json  -add-options json=omitempty  -transform=camelcase -w -file event.go
*/

// event raised after execution of a cli command
type CliCommand struct {
	Timestamp int64    `short:"t,omitempty" json:"timestamp,omitempty"`
	Duration  int64    `short:"d,omitempty" json:"duration,omitempty"`
	Version   string   `short:"v,omitempty" json:"version,omitempty"`
	Command   string   `short:"c,omitempty" json:"command,omitempty"`
	Args      []string `short:"a,omitempty" json:"args,omitempty"`
	Device    struct {
		OS        string `short:"o,omitempty" json:"os,omitempty"`
		ARCH      string `short:"h,omitempty" json:"arch,omitempty"`
		Username  string `short:"u,omitempty" json:"username,omitempty"`
		MachineID string `short:"m,omitempty" json:"machineID,omitempty"`
	} `short:"m,omitempty" json:"device,omitempty"`
	Workspace struct {
		Name  string `short:"n,omitempty" json:"name,omitempty"`
		Nodes int    `short:"o,omitempty" json:"nodes,omitempty"`
	} `short:"w,omitempty" json:"workspace,omitempty"`
	Project struct {
		Name        string `short:"n,omitempty" json:"name,omitempty"`
		Stages      int    `short:"s,omitempty" json:"stages,omitempty"`
		Nodes       int    `short:"o,omitempty" json:"nodes,omitempty"`
		AWSAccounts int    `short:"a,omitempty" json:"awsAccounts,omitempty"`
	} `short:"p,omitempty" json:"project,omitempty"`
	Stage struct {
		Name          string `short:"n,omitempty" json:"name,omitempty"`
		Node          string `short:"o,omitempty" json:"node,omitempty"`
		Functions     int    `short:"f,omitempty" json:"functions,omitempty"`
		PublicFolders int    `short:"p,omitempty" json:"publicFolders,omitempty"`
	} `short:"s,omitempty" json:"stage,omitempty"`
	Errors []CliError `short:"r,omitempty" json:"errors,omitempty"`
	Events []Event    `short:"e,omitempty" json:"events,omitempty"`
}

type CliError struct {
	Error        string `short:"e,omitempty" json:"error,omitempty"`
	Type         string `short:"t,omitempty" json:"type,omitempty"`
	SourceFile   string `short:"s,omitempty" json:"sourceFile,omitempty"`
	FunctionName string `short:"f,omitempty" json:"functionName,omitempty"`
}

func (c *CliCommand) Marshal() ([]byte, error) {
	buf, err := shortConfig.Marshal(c)
	if err != nil {
		return nil, err
	}
	return gz.Zip(buf)
}

func (c *CliCommand) Unmarshal(buf []byte) error {
	buf, err := gz.Unzip(buf)
	if err != nil {
		return err
	}
	return shortConfig.Unmarshal(buf, c)
}

func (c *CliCommand) Pretty() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

func (c *CliCommand) JSON() ([]byte, error) {
	return json.Marshal(c)
}

func (c *CliCommand) Add(e Event) {
	e.Timestamp = nowMS()
	c.Events = append(c.Events, e)
}

func (c *CliCommand) AddError(e CliError) {
	c.Errors = append(c.Errors, e)
}

func (c *CliCommand) Clear() {
	c.Events = make([]Event, 0)
	c.Errors = make([]CliError, 0)
}

// placeholder for all events
// only one attribute is not nil
type Event struct {
	Timestamp  int64       `short:"t,omitempty" json:"timestamp,omitempty"`
	GoBuild    *GoBuild    `short:"g,omitempty" json:"goBuild,omitempty"`
	Deploy     *Deploy     `short:"d,omitempty" json:"deploy,omitempty"`
	Signal     *Signal     `short:"s,omitempty" json:"signal,omitempty"`
	NodeCreate *NodeEvent  `short:"nc,omitempty" json:"nodeCreate,omitempty"`
	NodeDelete *NodeEvent  `short:"nd,omitempty" json:"nodeDelete,omitempty"`
	ProjectNew *ProjectNew `short:"n,omitempty" json:"projectNew,omitempty"`
	WatchCycle *WatchCycle `short:"wc,omitempty" json:"watchCycle,omitempty"`
	WatchDone  *WatchDone  `short:"wd,omitempty" json:"watchDone,omitempty"`
}

type GoBuild struct {
	Name     string `short:"n,omitempty" json:"name,omitempty"`
	Duration int    `short:"d,omitempty" json:"duration,omitempty"`
	Size     int    `short:"s,omitempty" json:"size,omitempty"`
}

type Deploy struct {
	Functions struct {
		Added   int `short:"a,omitempty" json:"added,omitempty"`
		Updated int `short:"u,omitempty" json:"updated,omitempty"`
		Removed int `short:"r,omitempty" json:"removed,omitempty"`
	} `short:"f,omitempty" json:"functions,omitempty"`
	PublicSites struct {
		Added   int `short:"a,omitempty" json:"added,omitempty"`
		Updated int `short:"u,omitempty" json:"updated,omitempty"`
		Removed int `short:"r,omitempty" json:"removed,omitempty"`
	} `short:"s,omitempty" json:"publicSites,omitempty"`
	InfrastructureChanged bool `short:"i,omitempty" json:"infrastructureChanged,omitempty"`
	BuildDuration         int  `short:"b,omitempty" json:"buildDuration,omitempty"`
	UploadDuration        int  `short:"u,omitempty" json:"uploadDuration,omitempty"`
	UploadBytes           int  `short:"m,omitempty" json:"uploadbytes,omitempty"`
	UpdateDuration        int  `short:"d,omitempty" json:"updateDuration,omitempty"`
}

type NodeEvent struct {
	AWSCredentialsProvider int    `short:"c,omitempty" json:"awsCredentialsProvider,omitempty"`
	StackDuration          int    `short:"s,omitempty" json:"stackDuration,omitempty"`
	InfrastructureDuration int    `short:"i,omitempty" json:"infrastructureDuration,omitempty"`
	AWSRegion              string `short:"r,omitempty" json:"region,omitempty"`
}

type ProjectNew struct {
	Name string `short:"n,omitempty" json:"name,omitempty"`
	From string `short:"f,omitempty" json:"from,omitempty"`
	Repo string `short:"r,omitempty" json:"repo,omitempty"`
}

type WatchDone struct {
	Cycles int `short:"c,omitempty" json:"cycles,omitempty"`
}

type WatchCycle struct {
	Duration   int  `short:"d,omitempty" json:"duration,omitempty"`
	CycleNo    int  `short:"c,omitempty" json:"cycleNo,omitempty"`
	HasUpdates bool `short:"b,omitempty" json:"hasUpdates,omitempty"`
	Invoke     bool `short:"i,omitempty" json:"invoke,omitempty"`
	Test       bool `short:"t,omitempty" json:"test,omitempty"`
}

const (
	AWSCredentialsByArguments = 1
	AWSCredentialsByEnv       = 2
	AWSCredentialsByProfile   = 3
)

type Signal struct {
	Name  string `short:"n,omitempty" json:"name,omitempty"`
	Stack string `short:"s,omitempty" json:"stack,omitempty"`
}

// marshal
var shortConfig = jsoniter.Config{
	EscapeHTML:             true,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
	TagKey:                 "short",
}.Froze()

func short(o interface{}) ([]byte, error) {
	return shortConfig.Marshal(o)
}

func NewCliCommand(buf []byte) (*CliCommand, error) {
	var cc CliCommand
	if err := cc.Unmarshal(buf); err != nil {
		return nil, err
	}
	return &cc, nil
}

func (c *CliCommand) Start() {
	c.Timestamp = nowMS()
	mid, err := machineid.ProtectedID("mantil")
	if err != nil {
		mid = "?"
	}
	u, _ := user.Current()
	c.Device.MachineID = mid
	c.Device.OS = runtime.GOOS
	c.Device.ARCH = runtime.GOARCH
	c.Device.Username = u.Username

	c.Args = RemoveAWSCredentials(os.Args)
	c.Version = Version()
}

func (c *CliCommand) End() {
	c.Duration = nowMS() - c.Timestamp
}

func nowMS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func RemoveAWSCredentials(args []string) []string {
	ak := regexp.MustCompile(`([A-Z0-9]){20}`)
	sak := regexp.MustCompile(`([a-zA-Z0-9+/]{40})`)

	for i, a := range args {
		a = sak.ReplaceAllString(a, "***")
		a = ak.ReplaceAllString(a, "***")
		args[i] = a
	}
	return args
}
