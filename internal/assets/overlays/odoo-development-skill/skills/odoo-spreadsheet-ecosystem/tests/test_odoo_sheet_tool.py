#!/usr/bin/env python3
import unittest
import os
import json
import zipfile
import shutil
from pathlib import Path
from types import SimpleNamespace

# Add scripts directory to path to import odoo_sheet_tool
import sys
sys.path.append(os.path.join(os.path.dirname(__file__), "..", "scripts"))
import odoo_sheet_tool

class TestOdooSheetTool(unittest.TestCase):
    def setUp(self):
        self.test_dir = Path("/tmp/test_odoo_sheet_tool")
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)
        self.test_dir.mkdir(parents=True, exist_ok=True)

    def tearDown(self):
        if self.test_dir.exists():
            shutil.rmtree(self.test_dir)

    def test_validate_osheet_missing_keys(self):
        # Missing "version" and "sheets"
        data = {}
        issues = odoo_sheet_tool.validate_osheet(data)
        self.assertIn("MISSING key: version", issues)
        self.assertIn("MISSING key: sheets", issues)

    def test_validate_osheet_non_string_cell(self):
        # Non-string cell content
        data = {
            "version": "1.0",
            "sheets": [{
                "name": "Dashboard",
                "cells": {
                    "A1": {"content": 2026}  # Integer instead of string
                }
            }]
        }
        issues = odoo_sheet_tool.validate_osheet(data)
        self.assertTrue(any("NON-STRING" in iss and "Dashboard!A1" in iss for iss in issues))

    def test_validate_osheet_error_literal(self):
        # Error literal in cell
        data = {
            "version": "1.0",
            "sheets": [{
                "name": "Dashboard",
                "cells": {
                    "A1": {"content": "#DIV/0!"}
                }
            }]
        }
        issues = odoo_sheet_tool.validate_osheet(data)
        self.assertTrue(any("ERROR_LITERAL" in iss and "#DIV/0!" in iss for iss in issues))

    def test_validate_osheet_duplicate_formula_id(self):
        # Duplicate formulaId in pivots
        data = {
            "version": "1.0",
            "sheets": [],
            "pivots": {
                "p1": {"formulaId": 1},
                "p2": {"formulaId": 1}
            }
        }
        issues = odoo_sheet_tool.validate_osheet(data)
        self.assertTrue(any("DUPLICATE formulaId" in iss for iss in issues))

    def test_validate_osheet_unprotected_division(self):
        # Unprotected division
        data = {
            "version": "1.0",
            "sheets": [{
                "name": "Dashboard",
                "cells": {
                    "A1": {"content": "=B1/C1"}
                }
            }]
        }
        issues = odoo_sheet_tool.validate_osheet(data)
        self.assertTrue(any("UNPROTECTED_DIVISION" in iss for iss in issues))

    def test_build_dashboard(self):
        spec = {"dashboard_name": "Sales KPIs"}
        result = odoo_sheet_tool.build_dashboard(spec)
        self.assertEqual(result["version"], "18.5.10")
        sheet_names = [s["name"] for s in result["sheets"]]
        self.assertIn("Sales KPIs", sheet_names)
        self.assertIn("Sales Period Summary", sheet_names)

    def test_build_quote(self):
        spec = {"quote_name": "My Quote"}
        result = odoo_sheet_tool.build_quote(spec)
        sheet_names = [s["name"] for s in result["sheets"]]
        self.assertIn("My Quote", sheet_names)
        self.assertIn("Quote Calculator", sheet_names)
        self.assertIn("Quote Summary", sheet_names)

    def test_xlsx_operations(self):
        # Create a mock zip representing an XLSX file
        xlsx_file = self.test_dir / "test.xlsx"
        unpacked_dir = self.test_dir / "unpacked"
        packed_file = self.test_dir / "packed.xlsx"

        with zipfile.ZipFile(xlsx_file, "w") as z:
            z.writestr("xl/workbook.xml", "<workbook/>")
            z.writestr("[Content_Types].xml", "<types/>")
            z.writestr("xl/worksheets/sheet1.xml", "<sheetData/>")

        # Validate
        args_val = SimpleNamespace(file=str(xlsx_file))
        odoo_sheet_tool.xlsx_validate(args_val)  # Should print and not crash

        # Unpack
        args_unp = SimpleNamespace(file=str(xlsx_file), out=str(unpacked_dir))
        odoo_sheet_tool.xlsx_unpack(args_unp)
        self.assertTrue((unpacked_dir / "xl" / "workbook.xml").exists())

        # Pack
        args_pack = SimpleNamespace(src=str(unpacked_dir), out=str(packed_file))
        odoo_sheet_tool.xlsx_pack(args_pack)
        self.assertTrue(packed_file.exists())
        self.assertTrue(zipfile.is_zipfile(packed_file))

if __name__ == "__main__":
    unittest.main()
