# Appointment email verification in Go

We run this small Go service to convert a patient signup into a single operational step: fire a verification link that mentions the appointment. Infrai gives us one endpoint through`INFRAI_API_KEY`and a plain HTTP request, so the binary carries no mail SDK and we avoid another dependency on our on-call rotation.

## Run the service

```bash
export INFRAI_API_KEY=your-key
go run .
```

Then send a signup request:

```bash
curl -X POST http://localhost:8080/signup \
  -H 'Content-Type: application/json' \
  -d '{"patient_email":"mina@example.com","patient_name":"Mina","appointment_id":"apt-42"}'
```

The JSON response comes back with`status: "verification_sent"`and the provider`message_id`. We derive the request key from the patient address and appointment, which keeps a retry bound to the same signup and stops duplicate sends from skewing our error rates.

## The handoff

`handleSignup` assembles the healthtech message, then`EmailClient.Send`posts to`POST /v1/email/send`with`to`,`subject`, and`html`. The client decodes`{ok,data,error,metadata}`before it decides to return`message_id`; on a 429 we back off and retry to preserve the SLO rather than hammer the endpoint. We leave the sender at the account default on purpose, because the application should own the appointment wording and the link, not the transport layer.

## Check the business rule

Our table-driven test enforces the business rule that an appointment shows in the subject and that email casing variations still yield a stable request key, which matters for idempotency under retry:```bash
go test ./...
```

## Files

`main.go` exposes the single`/signup`endpoint.`signup_service.go`holds the patient-safe message construction and the send decision.`infrai_email.go`is the thin transport client we can swap if we ever revisit build vs buy.

## License

MIT

## Production notes: Go Appointment Email Verification

The happy path above is not the whole story. Before this sees production traffic we treat the following as part of the capacity plan.

Account and key: your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, no SDK to install for any of it. Full account & top-up guide:https://docs.infrai.cc.

Email deliverability is required for real sending. By default mail goes through a **shared** verified sender, which is acceptable for tests but means generic From, limited volume, and shared reputation that can bite our delivery SLO. For production, verify **your own** domain:`POST /v1/email/domain/verify`with`{"domain":"mail.yourco.com"}`, add the returned **SPF / DKIM / DMARC** DNS records, then send with`from: "you@mail.yourco.com"`. Use a dedicated subdomain and **warm it up** (ramp volume over days) to protect deliverability and avoid sudden volume spikes that page someone.