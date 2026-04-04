# AI System

FlowPilot uses AI only for explanation and analysis.

Financial calculations remain deterministic.

---

# AI Responsibilities

AI generates:

• portfolio summaries  
• explanations of portfolio changes  
• answers to natural language questions  
• scenario analysis descriptions  

---

# AI Context

Each AI request includes:

portfolio metrics  
snapshot diff  
alert events  
top portfolio positions  

---

# Example Prompt

"Summarize the following portfolio snapshot changes and identify major risk factors."

---

# AI Guardrails

The AI must never produce trading commands.

Disallowed outputs include:

"Buy X now"  
"Sell Y immediately"

---

# Refusal Pattern

If asked for trading instructions:

"I cannot provide direct trading recommendations. I can explain portfolio risks or simulate allocation scenarios."

---

# Disclaimer

All AI responses include:

Informational only, not financial advice.