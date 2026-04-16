# Skill: Odoo AI Server Actions (v19)

## 1. Description
This skill defines architectural mandates for Odoo 19 native AI integration using the `ai.agent` framework. It ensures agents utilize the internal server action tool-calling mechanism instead of external, non-native APIs.

## 2. Core Mandatory Rules
- **Native Agents**: Only use `ai.agent` models. Never hallucinate third-party LLM service integrations.
- **Tool Calling**: Server Actions used as AI tools MUST set `use_in_ai=True`.
- **Result Injection**: NEVER use `return` statements in Server Actions. Use `ai['result'] = ...` to return data to the LLM.

## 3. Implementation Patterns (Bad vs Good)

### A. Exposing Server Actions as AI Tools
```python
# ❌ LEGACY: External API approach / Returning data
def get_info(self):
    return "Some data"

# ✅ v19 STANDARD: Native AI Server Action
# XML Definition: <field name="use_in_ai" eval="True"/>
# Python Code:
ai['result'] = f"Info: {self.get_info()}"
```

### B. AI Computed Fields
```python
# ❌ BAD: Manual logic in Python
def compute_summary(self):
    self.summary = openai.create_summary(self.description)

# ✅ v19 STANDARD: Declarative AI Automation
# Server Action Field:
# <field name="evaluation_type">ai_computed</field>
# <field name="ai_update_prompt">Summarize this text: /field</field>
```

## 4. Verification Workflow
- Ensure all AI-exposed actions define a JSON `ai_tool_schema` for input validation.
- Verify that `ai['result']` is populated in every tool-called method.

## 5. Maintenance
- Sync agent prompts with business requirements in `ai.topic`.
- Validate tool schema changes against native Odoo 19 `ai_server_actions` modules.
