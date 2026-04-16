# Skill: Odoo ORM Extreme Performance (v19)

## 1. Description
The `odoo-orm-extreme-19` skill defines the architectural mandates for ORM efficiency in Odoo 19. It enforces patterns that eliminate N+1 queries and high memory consumption.

## 2. Core Mandatory Rules
- **Indexing**: Composite and complex indexes **MUST** be defined declaratively via `models.Index`.
- **Query Efficiency**: `read_group()` is deprecated for backend logic. Use `_read_group()`.
- **Prefetching**: Use `search_fetch()` to perform search and prefetch in a single database round-trip.

## 3. Implementation Patterns (Bad vs Good)

### A. Declarative Indexing
```python
# ❌ LEGACY: Field-level index=True
name = fields.Char(index=True)

# ✅ v19 STANDARD: Declarative Index
class MyModel(models.Model):
    _indexes = [
        models.Index(name='custom_idx', expressions=['company_id', 'state'])
    ]
```

### B. Aggregation (The Performance Standard)
```python
# ❌ LEGACY: Expensive UI-metadata aggregation
results = self.env['model'].read_group(domain, ['amount:sum'], ['partner_id'])

# ✅ v19 STANDARD: Efficient Tuple Aggregation
results = self.env['model']._read_group(
    domain, 
    groupby=['partner_id'], 
    aggregates=['amount:sum']
)
```

## 4. Verification Workflow
- Ensure all custom modules replace `read_group` with `_read_group`.
- Audit database usage to ensure composite search queries use `_indexes`.

## 5. Maintenance
- Regularly audit performance-critical methods using `search_fetch()`.
- Keep the `ORM Performance Guide` in sync with Odoo 19 `models.py` structural changes.
