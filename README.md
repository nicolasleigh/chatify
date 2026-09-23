# 💬 Chatify

A **real-time chat app** with support for **audio and video calls**, rebuilt with **Golang** and **PostgreSQL** for high performance, efficient database operations, and scalable backend architecture.

## 📌 Project Overview

**Chatify** is a real-time messaging platform that enables users to chat, make audio and video calls, and manage conversations in a fast and scalable environment. While the initial version was built with **Next.js** and [**Convex**](https://www.convex.dev/), I later **rebuilt the backend using Golang and PostgreSQL** to address performance bottlenecks and gain full control over data management.

## 🎯 Motivation

What began as a course project became a deep-dive into backend optimization and scalability. The original implementation faced:

* Poor performance under scale
* Excessive database bandwidth usage
* N+1 query problems due to the limitations of Convex (lack of joins)

Rebuilding the backend allowed me to architect a **custom solution with raw SQL, efficient query handling, and real-time performance**, turning a prototype into a **production-grade application**.

## 🧑‍💻 My Role

I was responsible for full-stack development and system architecture:

* ✅ Rewrote the backend entirely in **Go**, integrating **PostgreSQL** for data persistence
* ⚙️ Designed efficient, raw **SQL queries** with **sqlc** for type-safe access
* 🔄 Built a custom **WebSocket** layer using **gorilla/websocket** for real-time messaging
* 🔐 Integrated **Clerk** authentication with webhook support to sync user data
* 🧱 Managed schema migrations with **golang-migrate**
* 🚀 Set up full **Docker-based deployment** with **Caddy** as reverse proxy
* 🧑‍🎨 Maintained and enhanced the existing **Next.js frontend**

## 🛠 Development Process

1. **Bottleneck Identification**
   * Database traffic hit 2GB/week
   * 400,000+ query calls from Convex
2. **Root Cause**
   * Convex backend lacked support for **SQL joins**, resulting in N+1 problems
3. **Redesign & Rebuild**
   * Switched to **Golang + PostgreSQL** with a focus on performance and control
4. **Backend Architecture**
   * Type-safe queries with [**sqlc**](https://sqlc.dev/)
   * Migrations managed via [**golang-migrate**](https://github.com/golang-migrate/migrate)
   * Real-time WebSocket logic using [**gorilla/websocket**](https://gorilla.github.io/) and Go routines
   * Live-reload with [**air**](https://github.com/air-verse/air)
   * Secret management with [**direnv**](https://direnv.net/)
5. **Deployment**
   * Dockerized all services and used **docker-compose**
   * Set up **Caddy** for automatic HTTPS and routing
6. **Authentication**
   * Integrated **Clerk**, with **webhook support** to sync user metadata

## 🔐 Required Runtime Configuration

The backend and frontend webhook handler must share the same high-entropy
`INTERNAL_WEBHOOK_SECRET`. The backend uses it to protect `POST /user` from
unauthorized identity synchronization requests, while the Next.js webhook
route sends it in the `X-Internal-Webhook-Secret` header.

Backend production configuration also supports:

* `CLOUD_DB_DSN` (or `DB_DSN` outside production)
* `CLERK_KEY`
* `RABBITMQ_URL`
* `RABBITMQ_EVENTS_EXCHANGE`
* `OUTBOX_POLL_INTERVAL_SECONDS`
* `OUTBOX_BATCH_SIZE`
* `NOTIFICATION_ENABLED`
* `NOTIFICATION_CONSUMER_QUEUE`
* `NOTIFICATION_POLL_INTERVAL_SECONDS`
* `NOTIFICATION_BATCH_SIZE`
* `NOTIFICATION_RETRY_LIMIT`
* `WEB_PUSH_VAPID_PUBLIC_KEY`
* `WEB_PUSH_VAPID_PRIVATE_KEY`
* `WEB_PUSH_VAPID_SUBJECT`
* `CORS_TRUSTED_ORIGINS` (comma-separated HTTP origins for browser and WebSocket access)
* `DB_MAX_CONNS` and `DB_MIN_CONNS`
* `DB_MAX_CONN_LIFETIME_SECONDS`
* `DB_MAX_CONN_IDLE_TIME_SECONDS`
* `DB_HEALTH_CHECK_PERIOD_SECONDS`

The matching `INTERNAL_WEBHOOK_SECRET` must also be present in the frontend
server environment. Do not expose it through a `NEXT_PUBLIC_*` variable.

Frontend deployments may override the backend endpoints with
`NEXT_PUBLIC_API_BASE_URL` and `NEXT_PUBLIC_WS_BASE_URL`.

RabbitMQ is used through a transactional PostgreSQL Outbox. Each HTTP or
WebSocket message write stores the chat message, updates the conversation
preview, and creates a `message.created` event in one database transaction. A
background worker then publishes the event to the durable RabbitMQ topic
exchange and only marks the Outbox row as published after a broker publisher
confirmation. Temporary failures are retried with exponential backoff; events
that exceed the retry limit remain in the database with status `failed` for
inspection and replay.

For Compose deployments, create these local secret files with one value per
file: `db_user.txt`, `db_password.txt`, `rabbitmq_user.txt`, and
`rabbitmq_password.txt`. Set `RABBITMQ_URL` in the deployment environment to
the matching private-network URL, for example
`amqp://user:password@rabbitmq:5672/`, then run `docker compose up --build`.
Do not commit any of those secret files.

## 🧰 Tech Stack

### 💻 Frontend

* **Next.js** – React framework for server-side rendering and routing
* **Tailwind CSS** – Utility-first styling

### 🖥 Backend

* **Golang** – High-performance server logic
* **sqlc** – Generate type-safe Go code from SQL
* **air** – Live reloading for efficient development
* **direnv** – Secure environment variable management

### 🗄 Database

* **PostgreSQL** – Relational database with schema migrations via `golang-migrate`

### 📨 Asynchronous Events

* **RabbitMQ** – Durable topic exchange with publisher confirms
* **Transactional Outbox** – Database-backed retry and failure retention

### 🔌 Real-Time & Auth

* **gorilla/websocket** – Real-time WebSocket communication
* **Clerk** – Authentication provider with webhook integration

### 🐳 Deployment

* **Docker** – Containerization for backend, frontend, and database
* **docker-compose** – Multi-service orchestration
* **Caddy** – Reverse proxy with automatic HTTPS

## ✨ Key Features

* 💬 Real-time chat with message persistence
* 📞 Audio & video call support
* ⚡️ Custom WebSocket backend in Go
* 🔒 Clerk-based authentication with webhook syncing
* 🧠 Type-safe, performant queries via raw SQL + sqlc
* 📦 Full Docker setup for local development and production deployment

## 🧗 Challenges & Solutions

| Problem                             | Solution                                                     |
| ----------------------------------- | ------------------------------------------------------------ |
| **Convex N+1 query issues**         | Switched to PostgreSQL and wrote optimized raw SQL with joins |
| **Socket.IO dependency on Node.js** | Built WebSocket logic from scratch using `gorilla/websocket` |
| **Backend iteration was slow**      | Used `air` for live-reloading, and `direnv` for env management |
