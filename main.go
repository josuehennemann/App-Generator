package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	//final dir is: dir/AppName
	ANDROID_FOLDER  = "android"
	IOS_FOLDER      = "ios"
	PROTOBUF_FOLDER = "proto"
	VIEW_FOLDER     = "view"
	SERVER_FOLDER   = "server"
	COMPANY_NAME    = "br.com.josuehennemann"
	SYSTEMD_FOLDER  = "systemd"

	URL_BASE_GIT   = "https://api.github.com/"
	AUTH_TOKEN_GIT = "9331815dd732aef94fee6f26319dd59e15c835e4"
)

var (
	flag_applicationName      string
	flag_ignoreFolders        string
	flag_gitRepo              bool
	alreadyCreatedGitRepo     bool
	allDirs                   map[string]*StContentDir
	RegexpNotAlpha            = regexp.MustCompile("[^a-zA-Z]")
	applicationNameNormalized string
)

func main() {

	flag.StringVar(&flag_applicationName, "name", "", "Application name. Ex:Personal Library")
	flag.BoolVar(&flag_gitRepo, "git", false, "Create git repository")
	flag.StringVar(&flag_ignoreFolders, "ignore-folders", "", "list of dirs that should not be created.Ex: server,android")

	flag.Parse()

	if flag_applicationName == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	//execute validations
	err := validateApplication()
	if err != nil {
		printError(err.Error())
	}
	
	if flag_gitRepo {
		err := createGitRepository()
		if err != nil {
			printError(err.Error())
		}
	}

	dirBase, err := os.Getwd()
	if err != nil {
		printError(err.Error())
	}
	applicationNameNormalized = normalizeAppName(flag_applicationName)
	setAllDirs()

	dirAppBase := dirBase + "/" + flag_applicationName
	createDir(dirAppBase)

	for d, content := range allDirs {
		folderPath := dirAppBase + "/" + d
		createDir(folderPath)
		for _, c := range content.List {
			if c.Dir {
				createDir(folderPath + "/" + c.Name)
				continue
			}

			createFile(folderPath+"/"+c.Name, c.Content)
		}
	}

	/*TODO:
	implemente validation method

	*/

}

type StContentDir struct {
	List []StItemContentDir
}

type StItemContentDir struct {
	Name    string
	Content string
	Dir     bool
}

func setAllDirs() {
	allDirs = map[string]*StContentDir{}

	allDirs[ANDROID_FOLDER] = buildContentDir(ANDROID_FOLDER) //&StContentDir{list:list}

	allDirs[IOS_FOLDER] = buildContentDir(IOS_FOLDER)           //&StContentDir{}
	allDirs[PROTOBUF_FOLDER] = buildContentDir(PROTOBUF_FOLDER) //&StContentDir{}
	allDirs[VIEW_FOLDER] = buildContentDir(VIEW_FOLDER)         //&StContentDir{}
	allDirs[SERVER_FOLDER] = buildContentDir(SERVER_FOLDER)     //&StContentDir{}

	if strings.TrimSpace(flag_ignoreFolders) != "" {
		for _, k := range strings.Split(flag_ignoreFolders, ",") {
			if _, ok := allDirs[k]; ok {
				delete(allDirs, k)
			}
		}
	}

}

func printError(s string) {
	fmt.Println(s)
	if alreadyCreatedGitRepo {
		fmt.Println("Excluindo repositorio no git")
		if err := deleteGitRepository(); err != nil {
			fmt.Println("Falha ao remover repositorio:", err.Error())
		}

	}
	os.Exit(1)
}

//TODO: implement
func validateApplication() error {
	return nil
}

func createDir(dir string) {
	err := os.Mkdir(dir, 0777)
	if err != nil {
		printError(err.Error())
	}
}

func createFile(file string, content string) {
	err := ioutil.WriteFile(file, []byte(content), 0777)
	if err != nil {
		printError(err.Error())
	}
}

func buildContentDir(tp string) *StContentDir {
	resp := &StContentDir{}
	list := []StItemContentDir{}
	itemReadme := StItemContentDir{Name: "README.md", Content: "Project " + tp}
	list = append(list, itemReadme)

	switch tp {
	case ANDROID_FOLDER:

	case IOS_FOLDER:

	case PROTOBUF_FOLDER:
		content := parseContentFile(contentFileProto)
		itemReadme := StItemContentDir{Name: applicationNameNormalized + ".proto", Content: content}
		list = append(list, itemReadme)
	case VIEW_FOLDER:

	case SERVER_FOLDER:
		content := parseContentFile(contentFileServerGo)
		itemReadme := StItemContentDir{Name: applicationNameNormalized + ".go", Content: content}
		list = append(list, itemReadme)
		folderSystemD := StItemContentDir{Name: SYSTEMD_FOLDER, Dir: true}
		list = append(list, folderSystemD)
		itemSystemD := StItemContentDir{Name: SYSTEMD_FOLDER + "/" + flag_applicationName + ".service", Content: "TODO: montar conteudo"}
		list = append(list, itemSystemD)

	default:
		return nil

	}

	resp.List = list

	return resp
}

func normalizeAppName(s string) string {
	//Upper Case first letter
	s = strings.Title(s)
	//remove not A-Z
	s = RegexpNotAlpha.ReplaceAllLiteralString(s, "")
	return s
}

func parseContentFile(content string) string {
	content = strings.Replace(content, "{{appname}}", flag_applicationName, -1)
	content = strings.Replace(content, "{{company}}", COMPANY_NAME, -1)
	content = strings.Replace(content, "{{appname_normalize}}", applicationNameNormalized, -1)
	content = strings.Replace(content, "{{year}}", fmt.Sprintf("%d", time.Now().Year()), -1)

	return content
}

type stAPIGitCreateRepoRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Private     bool   `json:"private"`
	Issues      bool   `json:"has_issues"`
	Projects    bool   `json:"has_projects"`
	Wiki        bool   `json:"has_wiki"`
	Init        bool   `json:"auto_init"`
}

func createGitRepository() error {
	url := URL_BASE_GIT + "user/repos"

	t := stAPIGitCreateRepoRequest{
		Name:        flag_applicationName,
		Description: "Projeto " + flag_applicationName,
		Init:        true,
	}

	b, err := json.Marshal(t)
	if err != nil {
		return err
	}

	payload := strings.NewReader(string(b))

	err = requestGit(http.MethodPost, url, payload, http.StatusCreated)
	if err != nil {
		return err
	}
	alreadyCreatedGitRepo = true
	return nil
}

func deleteGitRepository() error {

	url := URL_BASE_GIT + "repos/josuehennemann/" + flag_applicationName

	err := requestGit(http.MethodDelete, url, nil, http.StatusNoContent)
	if err != nil {
		return err
	}
	return nil
}

func requestGit(verbHttp, url string, payload io.Reader, statusHttp int) error {

	req, err := http.NewRequest(verbHttp, url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+AUTH_TOKEN_GIT)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != statusHttp {
		return fmt.Errorf("Falha na chamada de API para o github[%s]. Codigo http [%d]. Body [%s]", url, resp.StatusCode, body)
	}
	return nil

}

var contentFileProto = `
// Copyright {{year}} {{company}} authors.

syntax = "proto3";

option java_multiple_files = true;
option java_package = "{{company}}.{{appname_normalize}}";
option java_outer_classname = "{{appname_normalize}}Proto";

package {{appname}};
// The service definition.
service {{appname_normalize}} {
   	//TODO:
   	//implement another methods
	
	//example
   	//rpc SayHello (HelloRequest) returns (HelloReply) {}

 }
 
//example
// The request
//message HelloRequest {
//	string name = 1;
//}

// The response
message HelloReply {
//  string message = 1;
//}
`

var contentFileServerGo = `
/*
	Copyright {{year}} {{company}} authors.
	Generato files protobuf
	protoc -I ../{{appname}} --go_out=plugins=grpc:../{{appname}} ../proto/{{appname}}.proto
*/
package main

import (
	//"log"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement {{appname_normalize}}Server.
type server struct{}

//Example function implements {{appname_normalize}}Server
//func (s *server) SayHello(ctx context.Context, in *HelloRequest) (*HelloReply, error) {
//	return &HelloReply{Message: "Hello " + in.Name}, nil
//}

func main() {

	startGrpcServer()
}

func startGrpcServer(){
		lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		os.Exit(1)
	}
	s := grpc.NewServer()
	Register{{appname_normalize}}Server(s, &server{})
	// Register reflection service on gRPC server.
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		os.Exit(1)
	}
}


`
