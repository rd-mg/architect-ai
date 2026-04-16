# Skill: Odoo v19 Compliance Checker

## 1. Description
The `odoo-v19-compliance-checker` is a static analysis tool that scans Odoo modules for Odoo 19 mandatory breaking changes. It acts as an automated "Compliance Gate" to prevent non-compliant code (legacy SQL patterns, deprecated XML, missing types) from entering the codebase.

## 2. Mandatory Rules (Compliance Gate)
- **SQL Security**: Raw `cr.execute()` calls with formatted strings are strictly prohibited. The `SQL()` builder is mandatory for all database interactions.
- **Python Typing**: All public methods and `create` methods MUST include type hints (`-> 'ModelName'`, `-> bool`, etc.).
- **XML Deprecation**: The `<tree>` tag is removed; `<list>` must be used. The `attrs` attribute is deprecated in favor of direct view attributes.
- **ORM Patterns**: `_sql_constraints` is removed; `models.Constraint` is mandatory for data integrity.

## 3. Implementation Patterns
The checker utilizes Python's `ast` module to walk the Abstract Syntax Tree and `lxml` for XML structure validation.

```python
# AST Analysis Example for SQL() enforcement
def check_sql_builder(self, node):
    if isinstance(node, ast.Call) and getattr(node.func, 'attr', '') == 'execute':
        if not any(isinstance(arg, ast.Call) and getattr(arg.func, 'attr', None) == 'SQL' for arg in node.args):
             self.add_violation(file, node.lineno, "Use SQL() builder")
```

## 4. Verification Workflow
1. Execute `python odoo-v19-compliance-checker.py [module_path]`.
2. Review the JSON report output.
3. If `status: FAIL`, address all `CRITICAL` violations before further review.

## 5. Maintenance
- Add new v19 deprecation patterns to the `check_xml_file` or `check_python_file` methods as Odoo releases updates.
- Keep the checker synchronized with the official `models.py` structural changes in Odoo 19.
