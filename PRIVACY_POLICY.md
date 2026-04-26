# Privacy Policy — Hikmah Bot (حكمة)

**Last updated: April 26, 2026**

---

## 1. Overview

Hikmah Bot ("the Bot") is an Islamic educational Discord bot that provides interactive quizzes and Quran radio streaming. This privacy policy explains what data the Bot collects, how it is used, and your rights regarding that data.

---

## 2. Data We Collect

### 2.1 Data collected and stored

The Bot stores the following data in its PostgreSQL database:

| Data | Purpose | Retention |
|------|---------|-----------|
| Discord User ID | Tracking quiz scores within a session | Session duration only — deleted when the session ends |
| Discord Username | Displaying scores on the leaderboard at the end of a quiz session | Session duration only — deleted when the session ends |
| Quiz scores (points per session) | Leaderboard display | Session duration only |

### 2.2 Data we do NOT collect

- We do not store message content
- We do not store voice data or audio
- We do not store IP addresses
- We do not store email addresses
- We do not store any personally identifiable information beyond what is listed in 2.1
- We do not share any data with third parties
- We do not use data for advertising or analytics

---

## 3. How Data Is Used

Data collected by the Bot is used exclusively to:

- Display correct answers after a quiz question
- Show a per-session leaderboard at the end of a quiz session
- Route voice connections to the correct Discord channel during radio streaming

All session data is held **in memory only** during an active quiz session and is permanently discarded when the session ends. No quiz session data is written to the database.

---

## 4. Data Storage and Security

- The Bot's database is hosted on [Render](https://render.com) in a secured PostgreSQL instance
- Database access is restricted to the Bot's server — no public access is allowed
- All database connections use SSL/TLS encryption (`sslmode=require`)
- No session data (scores, user IDs) is persisted to the database after a session ends

---

## 5. Third-Party Services

The Bot interacts with the following external services:

| Service | Purpose | Their Privacy Policy |
|---------|---------|---------------------|
| Discord | Bot platform and user authentication | [discord.com/privacy](https://discord.com/privacy) |
| qurango.net | Live Quran radio stream URLs | External service — no user data is sent |
| Render | Bot hosting and database | [render.com/privacy](https://render.com/privacy) |

The Bot does not send any user data to any of these services beyond what Discord itself provides as part of normal bot operation (e.g. responding to interactions).

---

## 6. Children's Privacy

The Bot is designed as an educational Islamic tool suitable for all ages. We do not knowingly collect any data from children beyond what is described in Section 2.1, and all such data is discarded at session end.

---

## 7. Your Rights

You have the right to:

- **Know** what data the Bot holds about you (see Section 2.1 — only active session data)
- **Request deletion** — since no data persists after a session, there is nothing to delete once your session ends
- **Opt out** — simply do not participate in quiz sessions if you do not wish any temporary data to be held

To make a data request, open an issue on this repository or contact the bot owner via Discord.

---

## 8. Changes to This Policy

We may update this privacy policy from time to time. The **Last updated** date at the top of this document will reflect any changes. Continued use of the Bot after changes are posted constitutes acceptance of the updated policy.

---

## 9. Contact

For privacy concerns or data requests:

- **GitHub Issues:** [github.com/abdooman21/hikmah-bot/issues](https://github.com/abdooman21/hikmah-bot/issues)
- **Discord:** Contact the bot owner directly on the server where the Bot is installed

---

## 10. Legal Basis (GDPR)

For users in the European Economic Area, the legal basis for processing the temporary session data described in Section 2.1 is **legitimate interest** (Article 6(1)(f) GDPR) — specifically, providing the quiz leaderboard feature that users actively request by participating in a session.

Since this data is never persisted beyond the session, no long-term data processing occurs.

---

*Hikmah Bot is an open-source project. You can review exactly what data is collected and how it is handled by reading the source code at [github.com/abdooman21/hikmah-bot](https://github.com/abdooman21/hikmah-bot).*
