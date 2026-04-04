# TASK — Build FlowPilot Finance

You are Claude Code operating as a senior software engineer.

Your task is to build a **portfolio automation and analytics platform** called **FlowPilot Finance**.

The system aggregates portfolio holdings from multiple sources, computes portfolio metrics, detects changes, triggers alerts, and produces explainable AI insights.

The system must use **deterministic calculations** for financial logic and **AI only for explanations and insights**.

---

# Core Objective

Create a working application capable of:

• aggregating portfolio holdings  
• snapshotting portfolio state  
• detecting portfolio changes  
• triggering alerts  
• answering natural language questions about the portfolio  
• running portfolio scenario simulations  

---

# Tech Stack

Backend: Go  
Frontend: React + Typescript  
Database: MongoDB Atlas  
AI: Claude via LangChain  
Deployment: Docker (optional)

---

# System Architecture

External Data Sources  
↓  
Connectors  
↓  
Normalization Layer  
↓  
Portfolio Metrics Engine  
↓  
Snapshot Engine  
↓  
Diff Engine  
↓  
Alert Engine  
↓  
LangChain AI Layer  
↓  
Frontend Dashboard  

---

# Core Data Model

Position fields:

symbol  
quantity  
mark_price  
market_value  
asset_type  
source  
account_id  
timestamp  

---

# Portfolio Sync Workflow

When user triggers **Sync Portfolio**:

1 Fetch holdings from connectors  
2 Normalize into unified Position format  
3 Store in MongoDB  
4 Create portfolio snapshot  
5 Compute diff vs previous snapshot  
6 Evaluate alert rules  
7 Generate AI portfolio digest  

---

# Alert Types

exposure_threshold  
position_change_percent  
concentration_threshold  
net_worth_change  

---

# Natural Language Queries

Users should be able to ask:

"What changed in my portfolio today?"

"Which asset contributes most to risk?"

"What happens if crypto exposure drops to 20%?"

---

# AI Guardrails

The AI may:

• summarize portfolio changes  
• explain risk exposures  
• answer questions about holdings  

The AI must **never produce direct buy or sell instructions**.

If asked for trading advice, the system should respond:

"I cannot provide direct trading instructions, but I can explain portfolio risks or compare allocation scenarios."

---

# Deliverable

A working application with:

• portfolio upload capability  
• portfolio metrics dashboard  
• alert system  
• AI explanation engine  
• scenario simulation system