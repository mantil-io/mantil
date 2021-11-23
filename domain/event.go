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
	Timestamp int64    `short:"t,omitempty" json:"timestamp"`
	Duration  int64    `short:"d,omitempty" json:"duration"`
	Version   string   `short:"v,omitempty" json:"version"`
	Args      []string `short:"a,omitempty" json:"args"`
	User      *CliUser `short:"u,omitempty" json:"user,omitempty"`
	Device    struct {
		OS        string `short:"o,omitempty" json:"os"`
		ARCH      string `short:"h,omitempty" json:"arch"`
		Username  string `short:"u,omitempty" json:"username"`
		MachineID string `short:"m,omitempty" json:"machineID"`
	} `short:"m,omitempty" json:"device,omitempty"`
	Workspace *CliWorkspace `short:"w,omitempty" json:"workspace,omitempty"`
	Project   *CliProject   `short:"p,omitempty" json:"project,omitempty"`
	Stage     *CliStage     `short:"s,omitempty" json:"stage,omitempty"`
	Errors    []CliError    `short:"r,omitempty" json:"errors,omitempty"`
	Events    []Event       `short:"e,omitempty" json:"events,omitempty"`
}

type CliUser struct {
	ID    string `short:"i,omitempty" json:"id"`
	Email string `short:"e,omitempty" json:"email"`
}

type CliProject struct {
	Name        string `short:"n,omitempty" json:"name"`
	Stages      int    `short:"s,omitempty" json:"stages"`
	Nodes       int    `short:"o,omitempty" json:"nodes"`
	AWSAccounts int    `short:"a,omitempty" json:"awsAccounts"`
	AWSRegions  int    `short:"r,omitempty" json:"awsRegions"`
}

type CliStage struct {
	Name      string `short:"n,omitempty" json:"name"`
	Node      string `short:"o,omitempty" json:"node"`
	Functions int    `short:"f,omitempty" json:"functions"`
}

type CliWorkspace struct {
	ID          string `short:"i,omitempty" json:"ID"`
	Name        string `short:"n,omitempty" json:"name"`
	Nodes       int    `short:"o,omitempty" json:"nodes"`
	Projects    int    `short:"p,omitempty" json:"projects"`
	Stages      int    `short:"s,omitempty" json:"stages"`
	Functions   int    `short:"f,omitempty" json:"functions"`
	AWSAccounts int    `short:"a,omitempty" json:"awsAccounts"`
	AWSRegions  int    `short:"r,omitempty" json:"awsRegions"`
}

type CliError struct {
	Error        string `short:"e,omitempty" json:"error"`
	Type         string `short:"t,omitempty" json:"type"`
	SourceFile   string `short:"s,omitempty" json:"sourceFile"`
	FunctionName string `short:"f,omitempty" json:"functionName"`
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
	e.Timestamp = NowMS()
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
	Timestamp  int64       `short:"t,omitempty" json:"timestamp"`
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
	Name     string `short:"n,omitempty" json:"name"`
	Duration int    `short:"d,omitempty" json:"duration"`
	Size     int    `short:"s,omitempty" json:"size"`
}

type Deploy struct {
	Functions struct {
		Added   int `short:"a,omitempty" json:"added"`
		Updated int `short:"u,omitempty" json:"updated"`
		Removed int `short:"r,omitempty" json:"removed"`
	} `short:"f,omitempty" json:"functions,omitempty"`
	InfrastructureChanged bool `short:"i,omitempty" json:"infrastructureChanged"`
	BuildDuration         int  `short:"b,omitempty" json:"buildDuration"`
	UploadDuration        int  `short:"u,omitempty" json:"uploadDuration"`
	UploadBytes           int  `short:"m,omitempty" json:"uploadbytes"`
	UpdateDuration        int  `short:"d,omitempty" json:"updateDuration"`
}

type NodeEvent struct {
	AWSCredentialsProvider int    `short:"c,omitempty" json:"awsCredentialsProvider"`
	StackDuration          int    `short:"s,omitempty" json:"stackDuration"`
	InfrastructureDuration int    `short:"i,omitempty" json:"infrastructureDuration"`
	AWSRegion              string `short:"r,omitempty" json:"region"`
}

type ProjectNew struct {
	Name string `short:"n,omitempty" json:"name"`
	From string `short:"f,omitempty" json:"from"`
	Repo string `short:"r,omitempty" json:"repo"`
}

type WatchDone struct {
	Cycles int `short:"c,omitempty" json:"cycles"`
}

type WatchCycle struct {
	Duration   int  `short:"d,omitempty" json:"duration"`
	CycleNo    int  `short:"c,omitempty" json:"cycleNo"`
	HasUpdates bool `short:"b,omitempty" json:"hasUpdates"`
	Invoke     bool `short:"i,omitempty" json:"invoke"`
	Test       bool `short:"t,omitempty" json:"test"`
}

const (
	AWSCredentialsByArguments = 1
	AWSCredentialsByEnv       = 2
	AWSCredentialsByProfile   = 3
)

type Signal struct {
	Name  string `short:"n,omitempty" json:"name"`
	Stack string `short:"s,omitempty" json:"stack"`
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
	c.Timestamp = NowMS()
	mid := MachineID()
	u, _ := user.Current()
	c.Device.MachineID = mid
	c.Device.OS = runtime.GOOS
	c.Device.ARCH = runtime.GOARCH
	c.Device.Username = u.Username

	c.Args = RemoveAWSCredentials(os.Args)
	c.Version = Version()
}

func (c *CliCommand) End() {
	c.Duration = NowMS() - c.Timestamp
}

func NowMS() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func RemoveAWSCredentials(args []string) []string {
	// make copy before modify
	as := make([]string, len(args))
	copy(as, args)

	ak := regexp.MustCompile(`([A-Z0-9]){20}`)
	sak := regexp.MustCompile(`([a-zA-Z0-9+/]{40})`)

	for i, a := range as {
		a = sak.ReplaceAllString(a, "***")
		a = ak.ReplaceAllString(a, "***")
		as[i] = a
	}
	return as
}

func MachineID() string {
	mid, err := machineid.ProtectedID("mantil")
	if err != nil {
		mid = "?"
	}
	return mid
}
