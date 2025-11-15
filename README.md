<h1 align="center">⚡ AutoFlow.AI — Smart Workflow Orchestrator</h1>

<p align="center">
  <img 
    src="https://github.com/user-attachments/assets/18d2303e-230c-4637-9910-c87168edfed4"
    alt="AutoFlow Banner"
    width="900"
  />
</p>

<p align="center">
  <strong>The modern, open-source workflow engine for developers building automated systems on Go, AWS, and React.</strong>
</p>

<p align="center">
  <a href="https://github.com/Davanesh/autoFlow">⭐ Star this repo</a> ·
  <a href="https://github.com/Davanesh/autoFlow/issues">📦 Contribute</a> ·
  <a href="https://github.com/Davanesh/autoFlow/issues/new?labels=bug&template=bug_report.md">🐞 Report Bug</a> ·
  <a href="https://github.com/Davanesh/autoFlow/issues/new?labels=feature&template=feature_request.md">✨ Request Feature</a>
</p>

---

## 🔥 What is AutoFlow.AI?
AutoFlow.AI is a **workflow automation engine** that lets you design, run, and manage workflows using:

- ⚙️ **Go microservices**
- ☁️ **AWS Lambda, Step Functions, ECS**
- 🧠 **AI-assisted optimization**
- 🎛️ **React drag-and-drop workflow builder**
- 📐 **Modular & production-ready architecture**

Think of it like:
> “If **Google Cloud Workflows**, **n8n**, and **AWS Step Functions** had a baby that looks cool and works locally too.”

---

## 🎥 Demo (UI Preview)

<p align="center">
  <strong>demo video</strong><br/><br/>
  <a href="https://github.com/user-attachments/assets/a8baeaa3-d274-4186-a4cb-ee4bdc328073">
    ▶️ Watch Demo Video
  </a>
</p>

---

## 🧠 Features
- 🧩 Drag-and-drop workflow builder  
- ⚡ Real-time workflow execution logs  
- 🔗 Node-based visual flow system  
- 🔐 JWT-secured backend in Go  
- ☁️ Native AWS integrations  
- 💬 AI suggestions for workflow optimization  
- 📚 Version-controlled flows  
- 🌍 Fully open-source, easy to extend

---

## 🏗️ Tech Stack

| Area | Tech |
|------|------|
| Frontend | React, Zustand, Tailwind |
| Backend | Go + Fiber (Chi support planned) |
| Cloud | AWS (Lambda, Step Functions, S3, ECS) |
| Infra | Terraform |
| Auth | JWT or Firebase |
| Database | MongoDB (DynamoDB optional) |

---

## 🚀 Getting Started

```bash
# 1. Clone repo
git clone https://github.com/Davanesh/autoFlow

# 2. Install backend deps
cd backend/orchestrator
go mod tidy
go run main.go

# 3. Start frontend
cd frontend
npm install
npm run dev

