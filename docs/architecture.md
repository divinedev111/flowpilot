# FlowPilot Architecture

FlowPilot is an **automation-driven financial analytics system**.

The architecture separates **financial computation** from **AI explanation** to maintain reliability.

---

# Data Flow

External Data Sources  
↓  
Connectors  
↓  
Normalization Layer  
↓  
Portfolio Engine  
↓  
Snapshot Engine  
↓  
Diff Engine  
↓  
Alert Engine  
↓  
AI Insight Layer  
↓  
Dashboard

---

# Backend Responsibilities

The backend performs:

• financial calculations  
• portfolio aggregation  
• workflow automation  
• alert evaluation  

These operations must always remain deterministic.

---

# AI Layer

LangChain orchestrates interaction with Claude.

AI receives **structured context** including:

portfolio metrics  
snapshot differences  
alert results  

AI produces explanations based on that data.

---

# Frontend

The React frontend provides:

• portfolio dashboard  
• exposure charts  
• alert configuration  
• AI chat interface  

---

# Security Principles

• API keys never exposed to frontend  
• MongoDB restricted via network rules  
• deterministic financial calculations