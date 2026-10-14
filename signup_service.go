package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Signup struct {
	PatientEmail string `json:"patient_email"`
	PatientName  string `json:"patient_name"`
	Appointment  string `json:"appointment_id"`
}

type SignupResult struct {
	Status    string `json:"status"`
	MessageID string `json:"message_id,omitempty"`
}

func verifySubject(s Signup) string {
	return fmt.Sprintf("Verify your appointment %s, %s", s.Appointment, s.PatientName)
}

func verificationHTML(s Signup) string {
	return fmt.Sprintf("<p>Hello %s,</p><p>Confirm your email to manage appointment %s.</p><p><a href=\"https://clinic.example/verify?email=%s\">Verify email</a></p>", s.PatientName, s.Appointment, s.PatientEmail)
}

func signupID(s Signup) string {
	h := sha256.Sum256([]byte(strings.ToLower(s.PatientEmail) + ":" + s.Appointment))
	return hex.EncodeToString(h[:])
}

func handleSignup(client *EmailClient, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method must be POST", http.StatusMethodNotAllowed)
		return
	}
	var in Signup
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.PatientEmail == "" || in.PatientName == "" || in.Appointment == "" {
		http.Error(w, "patient_email, patient_name, and appointment_id are required", http.StatusBadRequest)
		return
	}
	id, err := client.Send(r.Context(), in.PatientEmail, verifySubject(in), verificationHTML(in), signupID(in))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(SignupResult{Status: "verification_sent", MessageID: id})
}

var _ = context.Background
