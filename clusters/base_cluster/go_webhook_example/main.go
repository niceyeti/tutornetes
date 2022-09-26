package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	// TODO: try this https://go.googlesource.com/tools/+/refs/heads/master/gopls/doc/workspace.md
	"go_webhook_example/src/mutator"

	admissionv1 "k8s.io/api/admission/v1"
)

const (
	CONTENT_TYPE = "Content-Type"
	JSON_CONTENT = "application/json"
)

func main() {
	// minimal handlers
	http.HandleFunc("/mutate-pods", ServeMutatePods)
	http.HandleFunc("/health", ServeHealth)

	// start the server
	if strings.ToLower(os.Getenv("TLS")) == "true" {
		key := os.Getenv("TLS_KEY_PATH")
		cert := os.Getenv("TLS_CERT_PATH")
		fmt.Println("Listening on port 443...")
		log.Fatal(http.ListenAndServeTLS(":443", cert, key, nil))
	} else {
		fmt.Println("Listening on port 8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

// ServeHealth returns 200 to signal liveness.
func ServeHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	//fmt.Fprint(w, http.StatusOK)
}

// ServeMutatePods serves the pod mutation endpoint.
func ServeMutatePods(w http.ResponseWriter, r *http.Request) {
	in, err := parseRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mt := mutator.PodMutator{
		Request: in.Request,
	}

	ar, err := mt.Mutate()
	if err != nil {
		e := fmt.Errorf("could not generate admission response: %v", err)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jout, err := json.Marshal(ar)
	fmt.Println("out: " + string(jout))
	if err != nil {
		e := fmt.Errorf("could not parse admission response: %v", err)
		http.Error(w, e.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", jout)
}

func parseRequest(r *http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get(CONTENT_TYPE) != JSON_CONTENT {
		return nil, fmt.Errorf(CONTENT_TYPE+": %q should be %q",
			r.Header.Get(CONTENT_TYPE), JSON_CONTENT)
	}

	bodybuf := new(bytes.Buffer)
	bodybuf.ReadFrom(r.Body)
	body := bodybuf.Bytes()

	if len(body) == 0 {
		return nil, fmt.Errorf("admission request body is empty")
	}

	var a admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("could not parse admission review request: %v", err)
	}

	if a.Request == nil {
		return nil, fmt.Errorf("admission review can't be used: Request field is nil")
	}

	return &a, nil
}
