#!/usr/bin/env python3
# -*- coding: utf-8 -*-

"""
Odoo o-spreadsheet JSON Builder (V2 - Strict)
Native generator for Odoo 19 spreadsheet data model.
Uses Pydantic V2 for structural validation with strict extra field forbidding.
"""

import json
import sys
import argparse
from typing import Dict, List, Optional, Any, Union
from pydantic import BaseModel, Field, ConfigDict

class BaseModelStrict(BaseModel):
    model_config = ConfigDict(extra="forbid")

class Style(BaseModelStrict):
    bold: Optional[bool] = None
    italic: Optional[bool] = None
    strikethrough: Optional[bool] = None
    underline: Optional[bool] = None
    fontSize: Optional[int] = None
    textColor: Optional[str] = None
    fillColor: Optional[str] = None
    verticalAlign: Optional[str] = None # top, middle, bottom
    textAlign: Optional[str] = None # left, center, right

class Cell(BaseModelStrict):
    content: str = "" # Formulas MUST start with =
    style: Optional[int] = None
    format: Optional[int] = None

class Header(BaseModelStrict):
    size: Optional[int] = None
    isHidden: Optional[bool] = None

class ConditionalFormat(BaseModelStrict):
    id: str
    ranges: List[str]
    rule: Dict[str, Any]

class Figure(BaseModelStrict):
    id: str
    type: str
    x: int
    y: int
    width: int
    height: int
    tag: Optional[str] = None
    data: Dict[str, Any]

class Pivot(BaseModelStrict):
    id: str
    type: str = "odoo"
    name: str
    model: str
    domain: List[Any] = []
    measures: List[Dict[str, str]] = []
    columns: List[str] = []
    rows: List[str] = []
    # Add other Odoo-specific pivot fields if known

class Sheet(BaseModelStrict):
    id: str
    name: str
    colNumber: int = 26
    rowNumber: int = 100
    cells: Dict[str, Cell] = {}
    merges: List[str] = []
    figures: List[Figure] = []
    conditionalFormats: List[ConditionalFormat] = []
    dataValidation: List[Dict[str, Any]] = []
    pivots: List[str] = [] # IDs of pivots used in this sheet
    cols: Dict[str, Header] = {}
    rows: Dict[str, Header] = {}

class OdooSpreadsheet(BaseModelStrict):
    version: int = 16
    sheets: List[Sheet]
    styles: Dict[int, Style] = {}
    formats: Dict[int, str] = {}
    borders: Dict[int, Any] = {}
    revisionId: str = "START_REVISION"

def create_minimal_spreadsheet(name="Sheet1"):
    sheet = Sheet(id="sheet1", name=name)
    return OdooSpreadsheet(sheets=[sheet])

def main():
    parser = argparse.ArgumentParser(description="Build Odoo o-spreadsheet JSON (Strict Mode)")
    parser.add_argument("--minimal", action="store_true", help="Print a minimal valid JSON")
    args = parser.parse_args()
    
    if args.minimal:
        ss = create_minimal_spreadsheet()
        print(ss.model_dump_json(indent=2))
        return

    # Example setup
    ss = create_minimal_spreadsheet()
    
    # Add an example cell with a formula (Odoo 19 style)
    ss.sheets[0].cells["A1"] = Cell(content="Hello Odoo")
    ss.sheets[0].cells["A2"] = Cell(content="=2+2")
    
    # Add a style
    ss.styles[1] = Style(bold=True, textColor="#FF0000")
    ss.sheets[0].cells["A1"].style = 1
    
    # Print the model
    try:
        print(ss.model_dump_json(indent=2))
    except Exception as e:
        print(f"Validation Error: {e}", file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
