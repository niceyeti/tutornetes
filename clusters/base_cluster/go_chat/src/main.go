package main

/*
For git and shiggles: a very basic web api to build as a 'webapp' for some helm deployment tutorials.
*/
import "fmt"

func main() {
	fmt.Println("starting super awesome app...")

	runHTTPServer()

	//port := getEnvOrDefault(envPort, defaultPort)
	//
	//r := mux.NewRouter()
	//r.HandleFunc("/", RootHandler).Methods("GET")
	//r.HandleFunc("/health", HealthHandler).Methods("GET")
	//r.HandleFunc("/fortune", FortuneHandler).Methods("GET")
	//r.HandleFunc("/echo", EchoHandler).Methods("POST")
	//
	//http.Handle("/", r)
	//log.Fatal(http.ListenAndServe(":"+port, r))
}
