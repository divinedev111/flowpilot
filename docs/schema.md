# MongoDB Schema

FlowPilot uses MongoDB Atlas.

---

# positions_current

Stores latest portfolio holdings.

Fields:

symbol  
quantity  
mark_price  
market_value  
asset_type  
source  
account_id  
timestamp  

---

# snapshots

Stores historical portfolio states.

Fields:

snapshot_id  
created_at  
net_worth  

---

# snapshot_positions

Positions belonging to each snapshot.

Fields:

snapshot_id  
symbol  
quantity  
mark_price  

---

# diffs

Portfolio changes between snapshots.

Fields:

from_snapshot  
to_snapshot  
net_worth_change  
top_movers  

---

# alert_rules

User-defined monitoring rules.

Fields:

rule_type  
threshold  
target_asset  

---

# alert_events

Triggered alerts.

Fields:

rule_id  
severity  
timestamp  
evidence  

---

# ai_digests

AI-generated summaries.

Fields:

snapshot_id  
summary  
risk_notes  
timestamp