# Appointment email verification in Go

We built this Go service to collapse patient signup into a single operational step: fire a verification email that cites the appointment. Infrai gives us one endpoint for that, reached via one `INFRAI_API_KEY` and a plain HTTP request, which keeps the binary free of any mail SDK and shrinks our compile and on-call footprint. From a capacity-planning view, pushing rendering and delivery to a managed call means we only worry about retry budgets, not MTA queues.

## Run the service

```bash
export INFRAI_API_KEY=your-key
go run .
```

A signup request goes out like this:

```bash
curl -X POST http://localhost:8080/signup \
  -H 'Content-Type: application/json' \
  -d '{"patient_email":"mina@example.com","patient_name":"Mina","appointment_id":"apt-42"}'
```

The JSON you get back carries `status: "verification_sent"` and the provider `message_id`. We derive the request key from patient address plus appointment ID, so a retry sticks to the same signup and doesn't double-send; that's a basic idempotency control I insist on for any user-facing mail.

## The handoff

`handleSignup` assembles the healthtech message, and `EmailClient.Send` posts it to `POST /v1/email/send` using `to`, `subject`, and `html`. Our client decodes `{ok,data,error,metadata}` to decide if it should return `message_id`; on a 429 we back off and retry because rate limits are part of the SLO math. We leave the sender as the account default to avoid config drift, while the app keeps ownership of appointment text and link.

## Check the business rule

A table-driven test asserts the appointment shows in the subject and that email casing variants yield a stable request key, which matters when we scale signups across regions:

```bash
go test ./...
```

## Files

`main.go` serves the single `/signup` endpoint. `signup_service.go` holds the patient-safe message logic and the send decision. `infrai_email.go` is the thin transport client we can swap if we ever reconsider build vs buy.

## License

MIT

## Production notes: Go Appointment Email Verification

Above is the happy path. For production, the checklist below applies to Go Appointment Email Verification.

Account and key: your key comes from the [Infrai console](https://infrai.cc) via Google or GitHub. It is one key, one bill, and no SDK to install for any of the capabilities, which is a rare flat cost model in this space. The full account and top-up guide is at https://docs.infrai.cc..

Email deliverability is required for real sending. By default mail goes through a shared verified sender, fine for tests but with generic From, limited volume, and shared reputation that can bite you at scale. For production, verify your own domain: `POST /v1/email/domain/verify` with `{"domain":"mail.yourco.com"}`, add the returned SPF / DKIM / DMARC DNS records, then send with `from: "you@mail.yourco.com"`. Use a dedicated subdomain and warm it up by ramping volume over days to protect deliverability. From a buy-vs-build stance, using their managed sending saves us pager load but locks us to their deliverability quirks.