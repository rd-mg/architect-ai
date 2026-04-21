# ripgrep-odoo — Local Codebase Discovery Skill (Odoo Monorepo)

## Propósito

Buscar patrones de código en el monorepo local de Odoo para obtener evidencia
real antes de diseñar o implementar. Usa el binario `rg` directamente.

**Usar SIEMPRE antes de:**
- Implementar cualquier funcionalidad que interactúe con el core de Odoo
- Asumir la API de un modelo, componente OWL, o acción de o-spreadsheet
- Escribir imports de módulos nativos de Odoo

**NO usar cuando:**
- Hay Engrams recientes con el patrón buscado (verificar primero con mem_search)
- La búsqueda es sobre código custom del proyecto (usar ripgrep base)

---

## Estructura del Monorepo (paths base)

```
~/gitproj/odoo/
├── odoo/            → core CE (modelos, controllers, vistas)
├── enterprise/      → módulos EE (si está disponible)
├── oca/             → módulos OCA (por subcarpeta según categoría)
├── o-spreadsheet/   → código fuente de o-spreadsheet
├── owl/             → framework OWL (si separado del core)
└── custom/          → módulos del proyecto actual (usar ripgrep base para esto)
```

---

## Dominios de Búsqueda y sus Flags

### backend_orm (modelos Python, ORM, lógica de negocio)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -t py \
  -g "!__manifest__.py" \
  -g "!__init__.py" \
  -g "!*/tests/*" \
  -g "!*/migrations/*" \
  --max-columns 150 \
  --max-count 3 \
  -C 3
```

Excluye: manifests, inits, tests, migrations. Solo lógica de negocio Python.

### frontend_owl (componentes OWL, JavaScript, XML de vistas)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/addons/web/static/src/ \
  -t js -t xml \
  -g "static/src/**" \
  -g "!*.min.js" \
  -g "!*.bundle.js" \
  --max-columns 150 \
  --max-count 3 \
  -C 2
```

Para un addon específico: sustituir `addons/web/static/src/` por `addons/{addon}/static/src/`

### o_spreadsheet (componentes de o-spreadsheet)

```bash
rg "{QUERY}" ~/gitproj/odoo/o-spreadsheet/src/ \
  -t js -t ts \
  -g "!*.min.js" \
  --max-columns 150 \
  --max-count 3 \
  -C 2
```

### views_xml (vistas Odoo, acciones, menús)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -t xml \
  -g "!static/" \
  -g "!*/i18n/*" \
  --max-columns 150 \
  --max-count 3 \
  -C 2
```

### security (ACLs, reglas de seguridad, grupos)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -g "security/**" \
  -g "*/ir.model.access.csv" \
  --max-count 5
```

### manifest (dependencias de módulos, version gates)

```bash
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -g "__manifest__.py" \
  --max-count 10
```

---

## Protocolo Obligatorio: 2 Pasos

**NUNCA ejecutar una búsqueda de contenido como primer paso.**

### Paso 1: Identificar archivos (-l, files-only)

```bash
# Primero: ¿en qué archivos está lo que busco?
rg "{QUERY}" ~/gitproj/odoo/community/ \
  -t py \           # (o el tipo del dominio)
  -l \              # solo nombres de archivos
  --max-count 5

# Ejemplo:
rg "class AccountMove" ~/gitproj/odoo/community/ -t py -l
# Output: 2-3 rutas de archivo (~10 tokens)
```

### Paso 2: Extracción quirúrgica del archivo relevante

```bash
# Solo si el Paso 1 devuelve el archivo correcto:
rg "{QUERY_ESPECÍFICA}" ~/gitproj/odoo/community/addons/account/models/account_move.py \
  --max-count 2 \
  -C 4

# Ejemplo:
rg "def _compute_payment_state" \
  ~/gitproj/odoo/community/addons/account/models/account_move.py \
  --max-count 2 -C 4
# Output: 15-20 líneas de código (~60 tokens)
```

**Por qué 2 pasos:** Paso 1 cuesta ~10 tokens. Paso 2 cuesta ~60 tokens. Total: 70 tokens.
Búsqueda directa sin filtrar: 40,000+ tokens → OOM.

---

## Banderas de Protección VRAM (SIEMPRE activas)

Estas banderas NUNCA deben omitirse en búsquedas sobre el monorepo completo:

| Flag | Valor | Por qué |
|---|---|---|
| `--max-columns` | `150` | Ignora líneas de archivos minificados (líneas de 80,000 chars) |
| `--max-count` | `3` (exploración) / `2` (extracción) | Limita resultados por archivo |
| `-C` | `2-4` (según dominio) | Contexto mínimo para entender el patrón |

Si el output supera 80 líneas o la búsqueda tarda más de 5 segundos:
→ STOP. Refinar con dominio más específico o subfolder más acotado.

---

## Patrones de Búsqueda por Caso de Uso

### "¿Cómo hereda este modelo en Odoo 18?"

```bash
# Paso 1:
rg "class AccountMove" ~/gitproj/odoo/community/ -t py -l

# Paso 2 (en el archivo encontrado):
rg "class AccountMove" ~/gitproj/odoo/community/addons/account/models/account_move.py -C 1
# → ve: class AccountMove(models.Model) o class AccountMove(account_move, models.Model)
```

### "¿Cómo importar este servicio en OWL?"

```bash
# Paso 1:
rg "useService" ~/gitproj/odoo/community/addons/web/static/src/ -t js -l --max-count 3

# Paso 2:
rg "const .* = useService" \
  ~/gitproj/odoo/community/addons/web/static/src/core/utils/hooks.js \
  --max-count 2 -C 2
```

### "¿Cómo agrega Odoo una acción a o-spreadsheet?"

```bash
# Paso 1:
rg "registry.category" ~/gitproj/odoo/community/addons/spreadsheet/static/src/ -t js -l

# Paso 2:
rg "registry.category\(\"spreadsheet" \
  ~/gitproj/odoo/community/addons/spreadsheet/static/src/ \
  -t js --max-count 2 -C 3
```

### "¿Qué ACL existe para este modelo?"

```bash
rg "account.bank.statement" \
  ~/gitproj/odoo/community/addons/account/ \
  -g "*/ir.model.access.csv" \
  --max-count 10
```

### "¿Qué módulos en OCA tienen este patrón?"

```bash
# Paso 1: solo archivos
rg "bank_statement" ~/gitproj/odoo/oca/ -t py -l --max-count 10

# Paso 2: pattern específico en el módulo relevante
rg "def _import_bank_statement" \
  ~/gitproj/odoo/oca/account-financial-tools/ \
  -t py --max-count 2 -C 3
```

---

## Persistencia al Engram (after_model hook)

Cuando ripgrep-odoo encuentra un patrón que resuelve una pregunta de diseño, el resultado DEBE persistirse al Engram para evitar buscar lo mismo en la próxima sesión. Además, si el patrón es de alta calidad y generalizable, debe proponerse su inclusión en el skill `patterns-{v}` correspondiente.

```
# Formato para mem_save después de una búsqueda exitosa:
topic_key: knowledge/odoo-v{N}/pattern/{slug-descriptivo}
content: {
  "query": "el rg command exacto que funcionó",
  "pattern": "el fragmento de código relevante (máx 20 líneas)",
  "source": "ruta relativa desde ~/gitproj/odoo/",
  "odoo_version": "18",  # o la versión confirmada
  "use_when": "descripción de cuándo usar este patrón",
  "propose_to_skill": true  # si debe agregarse a patterns-{v}
}
```

Este guardado lo ejecuta el **after_model hook** del Orchestrator (definido en Paso 05). Si `propose_to_skill` es `true`, el Orchestrator generará un artifact de tipo `task` para que un sub-agente actualice el `SKILL.md` de la versión correspondiente.

---

## Anti-patterns (qué NO hacer)

```bash
# ❌ NUNCA — búsqueda sin dominio en monorepo completo
rg "account_move" ~/gitproj/odoo/

# ❌ NUNCA — sin --max-count en búsqueda de contenido
rg "def compute" ~/gitproj/odoo/community/ -t py

# ❌ NUNCA — término demasiado genérico sin archivo específico
rg "def " ~/gitproj/odoo/community/addons/account/ -t py

# ✅ SIEMPRE — 2 pasos: files-only primero
rg "class BankStatementLine" ~/gitproj/odoo/community/ -t py -l
rg "_compute_amount" {archivo_del_paso_1} --max-count 2 -C 3
```
