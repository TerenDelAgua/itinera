# Privacy Policy — /privacy

**EN version (governing language: ES — see /privacy?lang=es)**
**Privacy Policy for Itinera**
**Last updated:** 1 September 2026 · **Version:** 1.0

---

## 1. Data Controller

| Field | Data |
|-------|------|
| Service | Itinera (a teren.dev project) |
| Contact email | privacy@teren.dev |
| Owner's website | https://teren.dev |

Itinera is a personal project operated under the teren.dev domain. No Data Protection Officer (DPO) has been appointed as it is not mandatory for this activity. For any privacy enquiry, please contact the email above.

## 2. What Data We Process

Itinera is a travel planning application. We process the following personal data:

| Category | Specific data | When |
|-----------|-----------------|--------|
| Identifiers | Email | When creating an account |
| Authentication | Password hash (bcrypt) | When creating an account |
| Technical | IP address, User-Agent | On every request (security and rate limiting) |
| Content | Trips, itineraries, expenses, notes | When using the app |
| Account metadata | Registration date, preferred language, tier | When creating an account |

What we do NOT do:
- We do not use tracking cookies or third-party analytics.
- We do not share data with advertisers.
- We do not sell data. Never.
- We do not use fingerprinting or advertising identifiers.

## 3. Purposes and Legal Basis

| Purpose | Legal basis (GDPR) |
|-----------|-------------------|
| Service provision (creating, editing, saving trips) | Art. 6.1.b — Performance of contract |
| Account management (registration, login, password recovery) | Art. 6.1.b — Performance of contract |
| Sending transactional emails (welcome, password reset) | Art. 6.1.b — Performance of contract |
| Security (rate limiting, abuse prevention) | Art. 6.1.f — Legitimate interest |
| Compliance with legal obligations | Art. 6.1.c — Legal obligation |

We do not carry out any processing based on consent. All processing is based on performance of contract or legitimate interest.

## 4. Recipients and Data Processors

We do not disclose data to third parties except under legal obligation. We use the following providers as data processors:

| Processor | Service | Location | Safeguards |
|-----------|----------|-----------|---------|
| Railway (Railway Corp.) | Backend hosting and database | USA / EU | Railway's standard DPA |
| Vercel (Vercel Inc.) | Frontend hosting | Global (CDN) | Vercel's standard DPA |
| Resend (Resend Inc.) | Transactional email sending | USA | Resend's standard DPA |

For transfers outside the EEA, we rely on the European Commission's Standard Contractual Clauses (SCCs) and/or the EU–US Data Privacy Framework.

## 5. Data Retention

| Data | Retention period |
|------|------------------|
| Active account | While the account exists |
| After deletion request | 30 days (soft delete) + permanent anonymisation |
| Trips of deleted users | Anonymised (linkage removed) |
| Inactive sessions | 30 days (automatic expiry) |
| Technical logs (IP, User-Agent) | Maximum 30 days |

## 6. Your Rights (Art. 15-22 GDPR)

You have the right to:
- **Access** — Obtain a copy of your data.
- **Rectification** — Correct inaccurate data.
- **Erasure** — Delete your account and all your data.
- **Restriction** — Temporarily restrict processing.
- **Portability** — Receive your data in a structured format (JSON).
- **Objection** — Object to processing based on legitimate interest.
- **Not be subject to automated decision-making with legal effects** — We do not make decisions based solely on automated processing that produce legal effects on you (Art. 22 GDPR). Automated security measures (rate limiting, abuse detection) do not produce legal effects.

**Identity verification:** we may request additional information to verify your identity before handling the request (Art. 12.6 GDPR). This is to protect your account.

How to exercise them:
- From the app: Settings → Delete account (immediate erasure).
- By email: Send your request to privacy@teren.dev stating which right you wish to exercise.

**Timeframes:** we respond within a maximum of 1 month from receipt of the request, extendable by 2 additional months in complex cases, notifying you of the extension and its reasons within the first month (Art. 12.3 GDPR). Exercising these rights is free of charge.

**Complaints:** if you believe we have not handled your request properly, you may lodge a complaint with the Spanish Data Protection Agency (AEPD): www.aepd.es

## 7. Security

We implement appropriate technical and organisational measures:
- Passwords hashed with bcrypt (cost 12).
- Sessions with rotating opaque tokens (not JWT).
- Rate limiting against brute force.
- Encrypted connections (HTTPS/TLS).
- Encryption at rest: AES-256 on database volumes.
- HttpOnly, Secure, SameSite cookies.
- No token exposure in URLs or logs.

In the event of a security breach affecting your data, we will notify you without undue delay and notify the AEPD within a maximum of 72 hours.

## 8. Cookies

Itinera only uses strictly necessary cookies for the operation of the service:

| Cookie | Purpose | Data stored | Duration |
|--------|-----------|---------------------|----------|
| `session_id` | Identify guest session | UUID v4 (no personal data) | 1 year |
| `itinera_access` | User authentication | Random opaque token (32 bytes). Stored hashed on the server, never in plain text. | 24 hours |
| `itinera_refresh` | Session renewal | Random opaque token (32 bytes, rotating). Stored hashed on the server, never in plain text. | 30 days |

We do not use third-party cookies, analytics, advertising or fingerprinting. As they are strictly necessary cookies, they do not require prior consent (Art. 22.2 LSSI).

## 9. Minors

Itinera is not directed at children under 14. We do not knowingly collect data from minors. If we discover that a minor has created an account, we will delete it.

## 10. Changes to This Policy

If we make material changes, we will notify you with reasonable advance notice (minimum 14 days) by in-app notification or email. The current version is the one published on this page with its last-updated date.

Previous versions are archived at `/privacy/archive` and can be consulted at any time.

## 11. Contact

For any privacy enquiry: privacy@teren.dev

## 12. Governing Law and Jurisdiction

These terms are governed by Spanish law and applicable European data protection regulations. For any dispute arising from this policy, the parties submit to the courts and tribunals of Valencia, Spain, without prejudice to the rights you have as a consumer in your place of habitual residence in the EEA.

---
