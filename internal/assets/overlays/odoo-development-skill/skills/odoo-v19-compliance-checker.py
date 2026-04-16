# -*- coding: utf-8 -*-

import ast
import json
import os
import sys
import lxml.etree as ET
from typing import Any, List, Dict

class Odoo19ComplianceChecker:
    def __init__(self, root_path: str):
        self.root_path = root_path
        self.report = {"violations": [], "status": "PASS"}

    def add_violation(self, file_path: str, line: int, message: str, level: str = "CRITICAL"):
        self.report["violations"].append({
            "file": file_path,
            "line": line,
            "message": message,
            "level": level
        })
        self.report["status"] = "FAIL"

    def check_python_file(self, file_path: str):
        with open(file_path, "r") as f:
            content = f.read()
            tree = ast.parse(content)

        for node in ast.walk(tree):
            # Check for raw SQL strings (cr.execute)
            if isinstance(node, ast.Call):
                if isinstance(node.func, ast.Attribute) and node.func.attr == 'execute':
                    if isinstance(node.func.value, ast.Attribute) and node.func.value.attr == 'cr':
                        # Check if arguments are NOT just a SQL object
                        if not (len(node.args) > 0 and isinstance(node.args[0], ast.Call) and 
                                (getattr(node.args[0].func, 'attr', None) == 'SQL' or 
                                 getattr(node.args[0].func, 'id', None) == 'SQL')):
                             self.add_violation(file_path, node.lineno, "Raw SQL detected. Use SQL() builder.")
            
            # Check for missing type hints in methods (ignoring __init__ for now)
            if isinstance(node, ast.FunctionDef) and node.name != '__init__':
                if not node.returns:
                    self.add_violation(file_path, node.lineno, f"Method {node.name} missing return type hint.")

    def check_xml_file(self, file_path: str):
        try:
            parser = ET.XMLParser(recover=True)
            tree = ET.parse(file_path, parser=parser)
            # 2.3 XML Deprecation Checks
            for node in tree.xpath('//tree'):
                self.add_violation(file_path, node.sourceline or 0, "Deprecated <tree> tag. Use <list>.")
            for node in tree.xpath('//*[@attrs]'):
                self.add_violation(file_path, node.sourceline or 0, "Deprecated 'attrs' attribute. Use direct attributes.")
        except Exception:
            pass

    def run(self):
        for root, _, files in os.walk(self.root_path):
            for file in files:
                path = os.path.join(root, file)
                if file.endswith('.py'):
                    self.check_python_file(path)
                elif file.endswith('.xml'):
                    self.check_xml_file(path)
        return self.report

if __name__ == "__main__":
    checker = Odoo19ComplianceChecker(sys.argv[1])
    report = checker.run()
    print(json.dumps(report, indent=2))
