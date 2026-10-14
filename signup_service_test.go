package main

import "testing"

func TestSignupDecision(t *testing.T) {
	tests := []struct {
		name        string
		in          Signup
		wantSubject string
		wantIDSame  bool
	}{
		{"appointment context", Signup{"a@example.com", "Mina", "apt-42"}, "Verify your appointment apt-42, Mina", true},
		{"case stable key", Signup{"A@EXAMPLE.COM", "Lee", "apt-42"}, "Verify your appointment apt-42, Lee", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verifySubject(tt.in); got != tt.wantSubject {
				t.Fatalf("subject = %q, want %q", got, tt.wantSubject)
			}
			if tt.wantIDSame && signupID(tt.in) != signupID(Signup{"a@example.com", tt.in.PatientName, tt.in.Appointment}) {
				t.Fatal("request key should be stable for an email address")
			}
		})
	}
}
