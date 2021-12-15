package main

/*
For git and shiggles: a very basic web api to build as a 'webapp' for some helm deployment tutorials.
*/
import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/gorilla/mux"
)

const (
	envPort     = "APP_PORT"
	defaultPort = "80"
)

/*
NOTE: This api is completely roughed out for a simple app api for the sake of cluster development.
Some basic info is printed to stdout, which using the default kubes logging will show up in app logs.
Otherwise nothing here is for quality. Before using, pay attention to the lib docs for http server
and gorillamux; especially the behavior of Write and WriteHeader:
	"If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK) before
	writing the data. ... yadda yadda ... side effects."
*/
func EchoHandler(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	fmt.Println("Echoing: " + string(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain")

	b = append(b, "\n"...)
	w.Write(b)
}

func FortuneHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Fortune hit")
	w.Header().Set("Content-Type", "text/plain")
	cmd := exec.Command("fortune")
	cmd.Stdout = w // this is efficient, but should determine any command errors before writing. this will write arbitrary failure data when non-200 responses are returned.
	if err := cmd.Run(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Health hit")
	/*
		Optional introspective health checks could go here...
	*/
	w.WriteHeader(http.StatusOK)
}

func RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Hit root")
	w.Header().Set("Content-Type", "text/html")
	http.FileServer(http.Dir("./static")).ServeHTTP(w, r)
}

func getEnvOrDefault(envVar, defaultVal string) string {
	val := os.Getenv(envVar)
	if val == "" {
		return defaultVal
	}
	//fmt.Println("Got " + val + " from var " + envVar)
	return val
}

func main() {
	fmt.Println("starting super awesome app...")

	port := getEnvOrDefault(envPort, defaultPort)

	r := mux.NewRouter()
	r.HandleFunc("/", RootHandler).Methods("GET")
	r.HandleFunc("/health", HealthHandler).Methods("GET")
	r.HandleFunc("/fortune", FortuneHandler).Methods("GET")
	r.HandleFunc("/echo", EchoHandler).Methods("POST")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
