package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	. "github.com/mattn/go-getopt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var lock = sync.RWMutex{}
var listen string

// The BCreate Type. Name of elements must match with jconf params
type BCreate struct {
	JName              string `json:"jname,omitempty"`
	XHCI               string `json:"xhci,omitempty"`
	AStart             string `json:"astart,omitempty"`
	RelativePath       string `json:"relative_path,omitempty"`
	Path               string `json:"path,omitempty"`
	Data               string `json:"data,omitempty"`
	RCConf             string `json:"rcconf,omitempty"`
	HostHostname       string `json:"host_hostname,omitempty"`
	IP4Addr            string `json:"ip4_addr,omitempty"`
	NicHWAddr          string `json:"nic_hwaddr,omitempty"`
	ZfsSnapSrc         string `json:"zfs_snapsrc,omitempty"`
	RunASAP            string `json:"runasap,omitempty"`
	Interface          string `json:"interface,omitempty"`
	RCtlNice           string `json:"rctl_nice,omitempty"`
	Emulator           string `json:"emulator,omitempty"`
	ImgSize            string `json:"imgsize,omitempty"`
	ImgType            string `json:"imgtype,omitempty"`
	VmCPUs             string `json:"vm_cpus,omitempty"`
	VmRAM              string `json:"vm_ram,omitempty"`
	VmOSType           string `json:"vm_os_type,omitempty"`
	VmEFI              string `json:"vm_efi,omitempty"`
	IsoSite            string `json:"iso_site,omitempty"`
	IsoImg             string `json:"iso_img,omitempty"`
	RegisterIsoName    string `json:"register_iso_name,omitempty"`
	RegisterIsoAs      string `json:"register_iso_as,omitempty"`
	VmHostBridge       string `json:"vm_hostbridge,omitempty"`
	BhyveFlags         string `json:"bhyve_flags,omitempty"`
	VirtioType         string `json:"virtio_type,omitempty"`
	VmOSProfile        string `json:"vm_os_profile,omitempty"`
	SwapSize           string `json:"swapsize,omitempty"`
	VmIsoPath          string `json:"vm_iso_path,omitempty"`
	VmGuestFS          string `json:"vm_guestfs,omitempty"`
	VmVNCPort          string `json:"vm_vnc_port,omitempty"`
	BhyveGenerateAcpi  string `json:"bhyve_generate_acpi,omitempty"`
	BhyveWireMemory    string `json:"bhyve_wire_memory,omitempty"`
	BhyveRtsKeepsUtc   string `json:"bhyve_rts_keeps_utc,omitempty"`
	BhyveForceMsiIrq   string `json:"bhyve_force_msi_irq,omitempty"`
	BhyveX2ApicMode    string `json:"bhyve_x2apic_mode,omitempty"`
	BhyveMpTableGen    string `json:"bhyve_mptable_gen,omitempty"`
	BhyveIgnoreMsrAcc  string `json:"bhyve_ignore_msr_acc,omitempty"`
	CdVncWait          string `json:"cd_vnc_wait,omitempty"`
	BhyveVNCResolution string `json:"bhyve_vnc_resolution,omitempty"`
	BhyveVNCTcpBind    string `json:"bhyve_vnc_tcp_bind,omitempty"`
	BhyveVNCVgaConf    string `json:"bhyve_vnc_vgaconf,omitempty"`
	NicDriver          string `json:"nic_driver,omitempty"`
	VNCPassword        string `json:"vnc_password,omitempty"`
	MediaAutoEject     string `json:"media_auto_eject,omitempty"`
	VmCPUTopology      string `json:"vm_cpu_topology,omitempty"`
	DebugEngine        string `json:"debug_engine,omitempty"`
	CdBootFirmware     string `json:"cd_boot_firmware,omitempty"`
	Jailed             string `json:"jailed,omitempty"`
	OnPowerOff         string `json:"on_poweroff,omitempty"`
	OnReboot           string `json:"on_reboot,omitempty"`
	OnCrash            string `json:"on_crash,omitempty"`
}

type Bhyves struct {
	JName    string
	JID      int
	VmRam    int // MB
	VmCPUs   int
	VmOSType string
	Status   string
	VNC      string
}

var bhyves []Bhyves

func init() {
	var c int
	// defaults
	listen = ":8080"

	OptErr = 0
	for {
		if c = Getopt("l:h"); c == EOF {
			break
		}
		switch c {
		case 'l':
			listen = OptArg
		case 'h':
			usage()
			os.Exit(1)
		}
	}
}

func usage() {
	println("usage: capi [-l listenaddress|-h]")
}

// main function to boot up everything
func main() {

	HandleInitBhyveList()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/blist", HandleBhyveList).Methods("GET")
	router.HandleFunc("/api/v1/cacheblist", HandleCacheBhyveList).Methods("GET")
	router.HandleFunc("/api/v1/bstart/{instanceid}", HandleBhyveStart).Methods("POST")
	router.HandleFunc("/api/v1/bstop/{instanceid}", HandleBhyveStop).Methods("POST")
	router.HandleFunc("/api/v1/bremove/{instanceid}", HandleBhyveRemove).Methods("POST")
	router.HandleFunc("/api/v1/bcreate/{instanceid}", HandleBhyveCreate).Methods("POST")
	log.Fatal(http.ListenAndServe(listen, router))
}

func HandleBhyveList(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	cmd := exec.Command("env", "NOCOLOR=0", "cbsd", "bls", "header=0", "display=jname,jid,vm_ram,vm_cpus,vm_os_type,status,vnc_port")
	stdout, err := cmd.Output()
	lock.Unlock()

	if err != nil {
		return
	}

	lines := strings.Split(string(stdout), "\n")
	imas := make([]Bhyves, 0)

	for _, line := range lines {
		if len(line) > 2 {
			ima := Bhyves{}
			reInsideWhtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			n := reInsideWhtsp.ReplaceAllString(line, " ")
			ima.JName = strings.Split(n, " ")[0]
			ima.JID, _ = strconv.Atoi(strings.Split(n, " ")[1])
			ima.VmRam, _ = strconv.Atoi(strings.Split(n, " ")[2])
			ima.VmCPUs, _ = strconv.Atoi(strings.Split(n, " ")[3])
			ima.VmOSType = strings.Split(n, " ")[4]
			ima.Status = strings.Split(n, " ")[5]
			ima.VNC = strings.Split(n, " ")[6]
			imas = append(imas, ima)
		}
	}

	_ = json.NewEncoder(w).Encode(&imas)
}

func HandleInitBhyveList() {
	lock.Lock()
	cmd := exec.Command("env", "NOCOLOR=0", "cbsd", "bls", "header=0", "display=jname,jid,vm_ram,vm_cpus,vm_os_type,status,vnc_port")
	stdout, err := cmd.Output()
	lock.Unlock()

	if err != nil {
		return
	}

	lines := strings.Split(string(stdout), "\n")

	for _, line := range lines {
		if len(line) > 2 {
			ima := Bhyves{}
			reInsideWhtsp := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			n := reInsideWhtsp.ReplaceAllString(line, " ")
			ima.JName = strings.Split(n, " ")[0]
			ima.JID, _ = strconv.Atoi(strings.Split(n, " ")[1])
			ima.VmRam, _ = strconv.Atoi(strings.Split(n, " ")[2])
			ima.VmCPUs, _ = strconv.Atoi(strings.Split(n, " ")[3])
			ima.VmOSType = strings.Split(n, " ")[4]
			ima.Status = strings.Split(n, " ")[5]
			ima.VNC = strings.Split(n, " ")[6]
			bhyves = append(bhyves, ima)
		}
	}
}

func HandleCacheBhyveList(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(&bhyves)
}

func HandleBhyveStart(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var instanceID string
	_ = json.NewDecoder(r.Body).Decode(&instanceID)
	instanceID = params["instanceid"]

	go realInstanceStart(instanceID)
}

func realInstanceStart(instanceID string) {
	jName := "jname=" + instanceID
	cmd := exec.Command("env", "NOCOLOR=0", "cbsd", "bstart", "inter=0", jName)
	_, err := cmd.Output()

	if err != nil {
		return
	}
}

func HandleBhyveStop(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var instanceID string
	buf, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		fmt.Printf("bodyErr %s", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	fmt.Printf("BODY rdr1: %q", rdr1)
	r.Body = rdr2

	_ = json.NewDecoder(r.Body).Decode(&instanceID)
	instanceID = params["instanceid"]

	go realInstanceStop(instanceID)
}

func realInstanceStop(instanceid string) {
	jName := "jname=" + instanceid
	cmd := exec.Command("env", "NOCOLOR=0", "cbsd", "bstop", "inter=0", jName)
	_, err := cmd.Output()

	if err != nil {
		return
	}
}

func HandleBhyveRemove(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	var instanceID string
	buf, bodyErr := ioutil.ReadAll(r.Body)

	if bodyErr != nil {
		//log.Print("bodyErr ", bodyErr.Error())
		fmt.Printf("bodyErr %s", bodyErr.Error())
		http.Error(w, bodyErr.Error(), http.StatusInternalServerError)
		return
	}

	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	fmt.Printf("BODY rdr1: %q", rdr1)
	//fmt.Printf("BODY rdr2: %q", rdr2)
	r.Body = rdr2

	_ = json.NewDecoder(r.Body).Decode(&instanceID)
	instanceID = params["instanceid"]
	go realInstanceRemove(instanceID)
}

func realInstanceRemove(instanceid string) {
	jName := "jname=" + instanceid
	cmd := exec.Command("env", "NOCOLOR=0", "cbsd", "bremove", "inter=0", jName)
	_, err := cmd.Output()

	if err != nil {
		return
	}
}

func realInstanceCreate(createstr string) {

	fmt.Printf("bcreate: [ %s ]", createstr)

	createstr = strings.TrimSuffix(createstr, "\n")
	arrCommandStr := strings.Fields(createstr)
	cmd := exec.Command(arrCommandStr[0], arrCommandStr[1:]...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func HandleBhyveCreate(w http.ResponseWriter, r *http.Request) {
	var instanceID string
	params := mux.Vars(r)
	instanceID = params["instanceid"]

	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}

	var bCreate BCreate
	_ = json.NewDecoder(r.Body).Decode(&bCreate)
	bCreate.JName = instanceID
	_ = json.NewEncoder(w).Encode(bCreate)
	val := reflect.ValueOf(bCreate)

	//	fmt.Println("J ",bCreate.JName)
	//	fmt.Println("R ",bCreate.VmRAM)

	var jConfParam string
	var str strings.Builder

	str.WriteString("env NOCOLOR=1 /usr/local/bin/cbsd bcreate inter=0 ")

	for i := 0; i < val.NumField(); i++ {
		//fmt.Printf("TEST %d\n",i);
		valueField := val.Field(i)

		typeField := val.Type().Field(i)
		tag := typeField.Tag

		tmpval := fmt.Sprintf("%s", valueField.Interface())

		if len(tmpval) == 0 {
			continue
		}

		jConfParam = strings.ToLower(typeField.Name)
		fmt.Printf("jconf: %s,\tField Name: %s,\t Field Value: %v,\t Tag Value: %s\n", jConfParam, typeField.Name, valueField.Interface(), tag.Get("tag_name"))
		buf := fmt.Sprintf("%s=%v ", jConfParam, tmpval)
		str.WriteString(buf)
	}

	go realInstanceCreate(str.String())
}
